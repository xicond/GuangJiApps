import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

/**
 * Converts an ArrayBuffer to a URL-safe Base64 string (RFC 4648 §5).
 */
export function bufferToBase64URL(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return btoa(binary)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

/**
 * Checks if the platform authenticator / WebAuthn is supported on the client device.
 */
export async function isPlatformAuthenticatorAvailable(): Promise<boolean> {
  if (typeof window === 'undefined') {
    return false
  }

  // Check WebAuthn support and platform authenticator availability
  if (window.PublicKeyCredential && typeof PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable === 'function') {
    try {
      const available = await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
      return available !== false
    } catch {
      return true
    }
  }

  // Fallback check against window.PublicKeyCredential
  return !(!window.PublicKeyCredential || !PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable)
}

export interface RegisterUserParams {
  id: string | number
  username: string
  name?: string
}

/**
 * Registers a new biometric WebAuthn passkey using navigator.credentials.create()
 * and posts the attestation payload to /v1/add-passkey.
 */
export async function registerBiometricPasskey(
  accessToken: string,
  user: RegisterUserParams
): Promise<string> {
  if (!navigator.credentials || !navigator.credentials.create) {
    throw new Error('WebAuthn tidak didukung pada browser ini.')
  }

  const userIdStr = String(user.id || '1')
  const userIdBytes = new TextEncoder().encode(userIdStr)
  const challengeBytes = new TextEncoder().encode(accessToken)

  const credential = (await navigator.credentials.create({
    publicKey: {
      challenge: challengeBytes,
      rp: {
        name: 'GuangJiApp'
      },
      user: {
        id: userIdBytes,
        name: user.username,
        displayName: user.name || user.username
      },
      pubKeyCredParams: [
        { type: 'public-key', alg: -7 } // ES256
      ],
      authenticatorSelection: {
        authenticatorAttachment: "platform", // "platform" memaksa biometrik lokal (TouchID/Windows Hello)
        userVerification: "required", // Wajib verifikasi sidik jari/wajah/PIN
        residentKey: "required", // Wajib disetel 'required' agar bisa dipakai untuk Autofill/Conditional UI
        requireResidentKey: true  // Fallback untuk browser versi lama
      }
    }
  })) as PublicKeyCredential | null

  if (!credential) {
    throw new Error('Pendaftaran biometrik dibatalkan atau gagal.')
  }

  const attestationResponse = credential.response as AuthenticatorAttestationResponse
  const payload = {
    id: credential.id,
    rawId: bufferToBase64URL(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64URL(attestationResponse.clientDataJSON),
      attestationObject: bufferToBase64URL(attestationResponse.attestationObject)
    }
  }

  const response = await axios.post(`${API_BASE}/v1/add-passkey`, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`
    }
  })

  return response.data?.refresh_token || ''
}

export interface RenewSessionResult {
  token: string
  refresh_token: string
}

/**
 * Prompts user for biometric assertion via navigator.credentials.get()
 * and exchanges it for a renewed access token and refresh token at /login-renew.
 */
export async function renewBiometricSession(
  refreshToken: string
): Promise<RenewSessionResult> {
  if (!navigator.credentials || !navigator.credentials.get) {
    throw new Error('WebAuthn tidak didukung pada browser ini.')
  }

  const challengeBytes = new TextEncoder().encode(refreshToken)

  const credential = (await navigator.credentials.get({
    mediation: "required",
    publicKey: {
      challenge: challengeBytes,
      // allowCredentials: [{
      //   id: credentialIdBuffer,
      //   type: "public-key",
      //   transports: ["internal"] // Memprioritaskan biometrik perangkat lokal
      // }],
      userVerification: 'required'
    }
  })) as PublicKeyCredential | null

  if (!credential) {
    throw new Error('Verifikasi biometrik dibatalkan.')
  }

  const assertionResponse = credential.response as AuthenticatorAssertionResponse
  const payload = {
    id: credential.id,
    rawId: bufferToBase64URL(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64URL(assertionResponse.clientDataJSON),
      authenticatorData: bufferToBase64URL(assertionResponse.authenticatorData),
      signature: bufferToBase64URL(assertionResponse.signature),
      userHandle: assertionResponse.userHandle ? bufferToBase64URL(assertionResponse.userHandle) : ''
    }
  }

  const response = await axios.post(`${API_BASE}/login-renew`, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${refreshToken}`
    }
  })

  const token = response.data?.token || response.data?.access_token || ''
  const newRefreshToken = response.data?.refresh_token || ''

  if (!token) {
    throw new Error('Access token baru tidak ditemukan pada response.')
  }

  return {
    token,
    refresh_token: newRefreshToken
  }
}
