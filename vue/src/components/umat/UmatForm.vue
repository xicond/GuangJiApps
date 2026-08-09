<template>
  <el-form ref="formRef" :model="formData" :rules="rules" label-position="top" size="default" class="umat-form"
    v-loading="submitting" @submit.prevent="handleSubmit">
    <el-alert v-if="props.fieldErrors && Object.keys(props.fieldErrors).length > 0" type="error" show-icon
      title="Invalid Inputs"
      description="Terdapat kesalahan pengisian form pada beberapa kolom di bawah ini. Silahkan periksa pesan kesalahan berwarna merah."
      class="validation-alert mb-4" />

    <!-- Section 1: Informasi Utama -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon>
            <User />
          </el-icon>
          <span>Informasi Utama</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Chiu Tao" prop="nama_indonesia" required
            :error="hasFieldError('nama_indonesia') ? ' ' : undefined">
            <el-input v-model="formData.nama_indonesia" placeholder="Masukkan nama Chiu Tao" />
            <FieldErrors :errors="getFieldErrors('nama_indonesia')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Alias / Pin Yin / Pin Yin" prop="alias"
            :error="hasFieldError('alias') ? ' ' : undefined">
            <el-input v-model="formData.alias" placeholder="Masukkan alias / Pin Yin" />
            <FieldErrors :errors="getFieldErrors('alias')" />
          </el-form-item>
        </el-col>

        <!-- <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Marga" prop="marga" :error="hasFieldError('marga') ? ' ' : undefined">
            <el-input v-model="formData.marga" placeholder="Masukkan marga" />
            <FieldErrors :errors="getFieldErrors('marga')" />
          </el-form-item>
        </el-col> -->

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Lain" prop="nama_mandarin"
            :error="hasFieldError('nama_mandarin') ? ' ' : undefined">
            <el-input v-model="formData.nama_mandarin" placeholder="Masukkan nama lain / panggilan" />
            <FieldErrors :errors="getFieldErrors('nama_mandarin')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Jenis Kelamin" prop="jenis_kelamin"
            :error="hasFieldError('jenis_kelamin') ? ' ' : undefined" required>
            <el-select v-model="formData.jenis_kelamin" placeholder="Pilih jenis kelamin" class="w-full">
              <el-option label="乾 PRIA" value="001" />
              <el-option label="坤 WANITA" value="002" />
              <el-option label="童 ANAK PRIA" value="003" />
              <el-option label="女 ANAK WANITA" value="004" />
            </el-select>
            <FieldErrors :errors="getFieldErrors('jenis_kelamin')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tempat Lahir" prop="tempat_lahir"
            :error="hasFieldError('tempat_lahir') ? ' ' : undefined">
            <el-input v-model="formData.tempat_lahir" placeholder="Tempat lahir" />
            <FieldErrors :errors="getFieldErrors('tempat_lahir')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Lahir" prop="tanggal_lahir"
            :error="hasFieldError('tanggal_lahir') ? ' ' : undefined" required>
            <el-date-picker v-model="formData.tanggal_lahir" type="date" placeholder="Pilih tanggal lahir"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('tanggal_lahir')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Pendidikan" prop="pendidikan" :error="hasFieldError('pendidikan') ? ' ' : undefined">
            <LookupSelect v-model="formData.pendidikan" placeholder="Pilih pendidikan"
              :fetch-api="lookupApi.getLookupPendidikan" auto-populate />
            <FieldErrors :errors="getFieldErrors('pendidikan')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Pekerjaan" prop="pekerjaan" :error="hasFieldError('pekerjaan') ? ' ' : undefined">
            <LookupSelect v-model="formData.pekerjaan" placeholder="Pilih pekerjaan"
              :fetch-api="lookupApi.getLookupPekerjaan" auto-populate />
            <FieldErrors :errors="getFieldErrors('pekerjaan')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Kode Buku" prop="kode_buku" :error="hasFieldError('kode_buku') ? ' ' : undefined">
            <el-input v-model="formData.kode_buku" placeholder="Masukkan kode buku (e.g. UMT-001)" />
            <FieldErrors :errors="getFieldErrors('kode_buku')" />
          </el-form-item>
        </el-col>

      </el-row>

      <el-row :gutter="16">
        <!-- <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tim Kerja" prop="tim_kerja" :error="hasFieldError('tim_kerja') ? ' ' : undefined">
            <LookupSelect v-model="formData.tim_kerja" placeholder="Pilih tim kerja"
              :fetch-api="lookupApi.getLookupTimKerja" />
            <FieldErrors :errors="getFieldErrors('tim_kerja')" />
          </el-form-item>
        </el-col> -->

        <el-col :xs="24">
          <el-form-item label="Keterangan" prop="keterangan" :error="hasFieldError('keterangan') ? ' ' : undefined">
            <el-input v-model="formData.keterangan" type="textarea" :rows="3" placeholder="Keterangan tambahan..." />
            <FieldErrors :errors="getFieldErrors('keterangan')" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>

    <!-- Section 3: Informasi Ciu Tao -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon>
            <Calendar />
          </el-icon>
          <span>Informasi Ciu Tao</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Ciu Tao (Masehi)" prop="tanggal_chiutao_int"
            :error="hasFieldError('tanggal_chiutao_int') ? ' ' : undefined" required>
            <el-date-picker v-model="formData.tanggal_chiutao_int" type="date" placeholder="Tanggal Ciu Tao Masehi"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('tanggal_chiutao_int')" />
          </el-form-item>
        </el-col>

        <!-- <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tanggal Ciu Tao (Man)" prop="tanggal_chiutao_man" :error="hasFieldError('tanggal_chiutao_man') ? ' ' : undefined">
            <el-input v-model="formData.tanggal_chiutao_man" placeholder="Tanggal Ciu Tao Mandarin" />
            <FieldErrors :errors="getFieldErrors('tanggal_chiutao_man')" />
          </el-form-item>
        </el-col> -->

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Waktu Ciu Tao" prop="waktu_chiutao_mandarin"
            :error="hasFieldError('waktu_chiutao_mandarin') ? ' ' : undefined" required>
            <LookupSelect v-model="formData.waktu_chiutao_mandarin" placeholder="Pilih waktu ciu tao"
              :fetch-api="lookupApi.getLookupWaktuCiuTao" :clearable="false" auto-populate />
            <FieldErrors :errors="getFieldErrors('waktu_chiutao_mandarin')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Fotang Ciu Tao" prop="fotang_chiutao"
            :error="hasFieldError('fotang_chiutao') ? ' ' : undefined" required>
            <LookupSelect v-model="formData.fotang_chiutao" placeholder="Pilih fotang ciu tao"
              :fetch-api="fotangApi.getFotangLookup" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('fotang_chiutao')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Tien Chuan Se (TCS)" prop="tcs" :error="hasFieldError('tcs') ? ' ' : undefined" required>
            <LookupSelect v-model="formData.tcs" placeholder="Pilih TCS" :fetch-api="lookupApi.getLookupTcs"
              :clearable="false" />
            <FieldErrors :errors="getFieldErrors('tcs')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Marga Tcs Lain" prop="nama_tcs_lain"
            :error="hasFieldError('nama_tcs_lain') ? ' ' : undefined">
            <el-input v-model="formData.nama_tcs_lain" placeholder="Marga Tcs Lain" />
            <FieldErrors :errors="getFieldErrors('nama_tcs_lain')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Uang Pahala" prop="uang_pahala" :error="hasFieldError('uang_pahala') ? ' ' : undefined">
            <el-input-number v-model="formData.uang_pahala" :min="10000" :step="1000" placeholder="0" class="w-full" />
            <FieldErrors :errors="getFieldErrors('uang_pahala')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Status Umat" prop="status_umat" :error="hasFieldError('status_umat') ? ' ' : undefined"
            required>
            <LookupSelect v-model="formData.status_umat" placeholder="Pilih status umat"
              :fetch-api="lookupApi.getLookupStatus" :clearable="false" auto-populate />
            <FieldErrors :errors="getFieldErrors('status_umat')" />
          </el-form-item>
        </el-col>


        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item v-if="formData.pengajak || !formData.pengajak_manual" label="Pengajak" prop="pengajak"
            :error="hasFieldError('pengajak') ? ' ' : undefined" required>
            <LookupSelect v-model="formData.pengajak" placeholder="Cari & pilih pengajak" :fetch-api="umatApi.getUmats"
              :get-item-api="umatApi.getUmatById" value-key="id" label-key="nama_indonesia"
              :label-formatter="formatUmatLabel" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('pengajak')" />
          </el-form-item>

          <el-form-item v-else label="Pengajak" prop="pengajak_manual"
            :error="hasFieldError('pengajak_manual') ? ' ' : undefined" required>
            <el-input v-model="formData.pengajak_manual" placeholder="Pengajak" />
            <FieldErrors :errors="getFieldErrors('pengajak_manual')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item v-if="formData.penanggung || !formData.penanggung_manual" label="Penanggung" prop="penanggung"
            :error="hasFieldError('penanggung') ? ' ' : undefined" required>
            <LookupSelect v-model="formData.penanggung" placeholder="Cari & pilih penanggung"
              :fetch-api="umatApi.getUmats" :get-item-api="umatApi.getUmatById" value-key="id"
              label-key="nama_indonesia" :label-formatter="formatUmatLabel" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('penanggung')" />
          </el-form-item>

          <el-form-item v-else label="Penanggung" prop="penanggung_manual"
            :error="hasFieldError('penanggung_manual') ? ' ' : undefined" required>
            <el-input v-model="formData.penanggung_manual" placeholder="Penanggung" />
            <FieldErrors :errors="getFieldErrors('penanggung_manual')" />
          </el-form-item>
        </el-col>

      </el-row>
    </el-card>

    <!-- Section 2: Kontak, Alamat & Fotang -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon>
            <Phone />
          </el-icon>
          <span>Kontak & Fotang</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Telepon" prop="telepon" :error="hasFieldError('telepon') ? ' ' : undefined">
            <el-input v-model="formData.telepon" placeholder="Nomor telepon" />
            <FieldErrors :errors="getFieldErrors('telepon')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Mobile" prop="mobile" :error="hasFieldError('mobile') ? ' ' : undefined">
            <el-input v-model="formData.mobile" placeholder="Nomor Mobile" />
            <FieldErrors :errors="getFieldErrors('mobile')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Email" prop="email" :error="hasFieldError('email') ? ' ' : undefined">
            <el-input v-model="formData.email" placeholder="Alamat email" type="email" />
            <FieldErrors :errors="getFieldErrors('email')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Alamat Domisili" prop="alamat" :error="hasFieldError('alamat') ? ' ' : undefined">
            <el-input v-model="formData.alamat" type="textarea" :rows="2" placeholder="Alamat domisili" />
            <FieldErrors :errors="getFieldErrors('alamat')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Alamat Lain" prop="alamat2" :error="hasFieldError('alamat2') ? ' ' : undefined">
            <el-input v-model="formData.alamat2" type="textarea" :rows="2" placeholder="Alamat lain" />
            <FieldErrors :errors="getFieldErrors('alamat2')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Wilayah" prop="wilayah" :error="hasFieldError('wilayah') ? ' ' : undefined">
            <el-input v-model="formData.wilayah" type="text" placeholder="Wilayah" />
            <FieldErrors :errors="getFieldErrors('wilayah')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Fotang Aktif" prop="fotang_aktif"
            :error="hasFieldError('fotang_aktif') ? ' ' : undefined" required>
            <LookupSelect v-model="formData.fotang_aktif" placeholder="Pilih fotang aktif"
              :fetch-api="fotangApi.getFotangLookup" :clearable="false" />
            <FieldErrors :errors="getFieldErrors('fotang_aktif')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item label="Nama Fotang Lain" prop="nama_fotang_lain"
            :error="hasFieldError('nama_fotang_lain') ? ' ' : undefined">
            <el-input v-model="formData.nama_fotang_lain" placeholder="Nama fotang lain" />
            <FieldErrors :errors="getFieldErrors('nama_fotang_lain')" />
          </el-form-item>
        </el-col>

      </el-row>
    </el-card>


    <!-- Section 4: Kelas & Sidang Dharma (Sd2 / Sd3) -->
    <el-card shadow="never" class="form-section-card">
      <template #header>
        <div class="section-title">
          <el-icon>
            <Notebook />
          </el-icon>
          <span>Kelas, Sidang Dharma, Dll</span>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item label="Kelas Umum" prop="kelas_umum" :error="hasFieldError('kelas_umum') ? ' ' : undefined">
            <LookupSelect v-model="formData.kelas_umum" placeholder="Pilih kelas umum"
              :fetch-api="lookupApi.getLookupKelasUmum" />
            <FieldErrors :errors="getFieldErrors('kelas_umum')" />
          </el-form-item>
        </el-col>

        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item label="Kelas Khusus" prop="kelas_khusus"
            :error="hasFieldError('kelas_khusus') ? ' ' : undefined">
            <LookupSelect v-model="formData.kelas_khusus" placeholder="Pilih kelas khusus"
              :fetch-api="lookupApi.getLookupKelas" />
            <FieldErrors :errors="getFieldErrors('kelas_khusus')" />
          </el-form-item>
        </el-col>

      </el-row>
      <el-row :gutter="16">

        <!-- Sidang Dharma Pemula (Sd3) -->
        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item prop="sd3"
            :error="hasFieldError('sd3') || hasFieldError('tanggal_sd3') || hasFieldError('tempat_sd3') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.sd3">Sidang Dharma Pemula</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_sd3" type="date" placeholder="Tanggal Sidang Dharma Pemula"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.sd3" />
            <LookupSelect v-model="formData.tempat_sd3" placeholder="Pilih tempat Sidang Dharma Pemula"
              :fetch-api="fotangApi.getFotangLookup" :disabled="!formData.sd3" class="w-full mt-2" />
            <FieldErrors
              :errors="getFieldErrors('sd3').concat(getFieldErrors('tanggal_sd3')).concat(getFieldErrors('tempat_sd3'))" />
          </el-form-item>
        </el-col>

        <!-- Fo Kuei Li Cie Pan (Sd2) -->
        <el-col :xs="24" :sm="12" :md="12">
          <el-form-item prop="sd2"
            :error="hasFieldError('sd2') || hasFieldError('tanggal_sd2') || hasFieldError('tempat_sd2') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.sd2">Fo Kuei Li Cie Pan</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_sd2" type="date" placeholder="Tanggal Fo Kuei Li Cie Pan"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.sd2" />
            <LookupSelect v-model="formData.tempat_sd2" placeholder="Pilih tempat Fo Kuei Li Cie Pan"
              :fetch-api="fotangApi.getFotangLookup" :disabled="!formData.sd2" class="w-full mt-2" />
            <FieldErrors
              :errors="getFieldErrors('sd2').concat(getFieldErrors('tanggal_sd2')).concat(getFieldErrors('tempat_sd2'))" />
          </el-form-item>
        </el-col>

      </el-row>

      <el-row :gutter="16">
        <!-- Ikrar Checkboxes -->
        <el-col :xs="24">
          <el-form-item label="Ikrar" prop="ikrar" :error="hasFieldError('ikrar_1') ? ' ' : undefined">
            <div class="ikrar-checkbox-group">
              <el-checkbox v-model="formData.ikrar_1" label="Ikrar 重聖輕凡" />
              <el-checkbox v-model="formData.ikrar_2" label="Ikrar 財法雙施" />
              <el-checkbox v-model="formData.ikrar_3" label="Ikrar 清口茹素" />
              <el-checkbox v-model="formData.ikrar_4" label="Ikrar 捨身辦道" />
              <el-checkbox v-model="formData.ikrar_5" label="Ikrar 開設佛堂" />
              <el-checkbox v-model="formData.ikrar_6" label="Ikrar 開荒下種" />
            </div>
            <FieldErrors
              :errors="getFieldErrors('ikrar_1').concat(getFieldErrors('ikrar_2')).concat(getFieldErrors('ikrar_3')).concat(getFieldErrors('ikrar_4')).concat(getFieldErrors('ikrar_5')).concat(getFieldErrors('ikrar_6'))" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <!-- Ren Chai Pan -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item prop="ren_chai_pan"
            :error="hasFieldError('ren_chai_pan') || hasFieldError('tanggal_ren_chai_pan') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.ren_chai_pan">Ren Chai Pan</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_ren_chai_pan" type="date" placeholder="Tanggal Ren Chai Pan"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.ren_chai_pan" />
            <FieldErrors :errors="getFieldErrors('ren_chai_pan').concat(getFieldErrors('tanggal_ren_chai_pan'))" />
          </el-form-item>
        </el-col>

        <!-- Lien Ciang Pan -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item prop="lien_ciang_pan"
            :error="hasFieldError('lien_ciang_pan') || hasFieldError('tanggal_lien_ciang_pan') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.lien_ciang_pan">Lien Ciang Pan</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_lien_ciang_pan" type="date" placeholder="Tanggal Lien Ciang Pan"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.lien_ciang_pan" />
            <FieldErrors :errors="getFieldErrors('lien_ciang_pan').concat(getFieldErrors('tanggal_lien_ciang_pan'))" />
          </el-form-item>
        </el-col>

        <!-- Ciang Yen Pan -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item prop="ciang_yen_pan"
            :error="hasFieldError('ciang_yen_pan') || hasFieldError('tanggal_ciang_yen_pan') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.ciang_yen_pan">Ciang Yen Pan</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_ciang_yen_pan" type="date" placeholder="Tanggal Ciang Yen Pan"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.ciang_yen_pan" />
            <FieldErrors :errors="getFieldErrors('ciang_yen_pan').concat(getFieldErrors('tanggal_ciang_yen_pan'))" />
          </el-form-item>
        </el-col>

        <!-- Ching Khou -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item prop="ching_khou"
            :error="hasFieldError('ching_khou') || hasFieldError('tanggal_ching_khou') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.ching_khou">Ching Khou</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_ching_khou" type="date" placeholder="Tanggal Ching Khou"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.ching_khou" />
            <FieldErrors :errors="getFieldErrors('ching_khou').concat(getFieldErrors('tanggal_ching_khou'))" />
          </el-form-item>
        </el-col>

        <!-- An Cuo -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item prop="tanggal_ancuo"
            :error="hasFieldError('tanggal_ancuo') || hasFieldError('nama_cetya_rumah') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="hasAnCuo">An Cuo</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_ancuo" type="date" placeholder="Tanggal An Cuo"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!hasAnCuo" />
            <el-form-item prop="nama_cetya_rumah" class="w-full mt-2" style="margin-bottom: 0;">
              <el-input v-model="formData.nama_cetya_rumah" placeholder="Nama Cetya Rumah" :disabled="!hasAnCuo"
                class="w-full" />
            </el-form-item>
            <FieldErrors :errors="getFieldErrors('tanggal_ancuo').concat(getFieldErrors('nama_cetya_rumah'))" />
          </el-form-item>
        </el-col>

        <!-- Meninggal -->
        <el-col :xs="24" :sm="12" :md="8">
          <el-form-item prop="meninggal"
            :error="hasFieldError('meninggal') || hasFieldError('tanggal_meninggal') ? ' ' : undefined">
            <template #label>
              <div @click.stop>
                <el-checkbox v-model="formData.meninggal">Sudah Meninggal</el-checkbox>
              </div>
            </template>
            <el-date-picker v-model="formData.tanggal_meninggal" type="date" placeholder="Tanggal Meninggal"
              format="YYYY-MM-DD" value-format="YYYY-MM-DD" class="w-full" :disabled="!formData.meninggal" />
            <FieldErrors :errors="getFieldErrors('meninggal').concat(getFieldErrors('tanggal_meninggal'))" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-card>


    <!-- Form Action Buttons -->
    <div class="form-actions">
      <el-button size="large" :disabled="submitting" @click="handleCancel">Batal</el-button>
      <el-button type="primary" size="large" :loading="submitting" :disabled="submitting" :icon="Check"
        @click="handleSubmit">
        {{ submitText }}
      </el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import dayjs from 'dayjs'
import type { FormInstance, FormRules } from 'element-plus'
import { User, Phone, Calendar, Notebook, Check } from '@element-plus/icons-vue'
import type { Umat } from '../../types/umat'
import LookupSelect from '../common/LookupSelect.vue'
import FieldErrors from '../common/FieldErrors.vue'
import lookupApi from '../../api/lookup'
import fotangApi from '../../api/fotang'
import umatApi from '../../api/umat'

const props = withDefaults(
  defineProps<{
    initialData?: Partial<Umat>
    submitting?: boolean
    submitText?: string
    fieldErrors?: Record<string, string[]>
  }>(),
  {
    initialData: () => ({}),
    submitting: false,
    submitText: 'Simpan Data',
    fieldErrors: () => ({})
  }
)

const emit = defineEmits<{
  (e: 'submit', payload: Partial<Umat>): void
  (e: 'cancel'): void
}>()

const formRef = ref<FormInstance>()

function getFieldErrors(fieldName: string): string[] {
  if (!props.fieldErrors) return []
  return props.fieldErrors[fieldName] || []
}

function hasFieldError(fieldName: string): boolean {
  return getFieldErrors(fieldName).length > 0
}

// Local reactive state for form fields
const formData = ref<Partial<Umat>>({
  kode_buku: '',
  nama_indonesia: '',
  alias: '',
  marga: '',
  nama_mandarin: '',
  jenis_kelamin: '',
  tempat_lahir: '',
  tanggal_lahir: null,
  pendidikan: '',
  pekerjaan: '',
  telepon: '',
  mobile: '',
  email: '',
  alamat: '',
  alamat2: '',
  fotang_aktif: '',
  nama_fotang_lain: '',
  tahun_chiutao_mandarin: '',
  waktu_chiutao_mandarin: '',
  tanggal_chiutao_int: null,
  tanggal_chiutao_man: '',
  pengajak: '',
  penanggung: '',
  tcs: '',
  nama_tcs_lain: '',
  uang_pahala: 0,
  fotang_chiutao: '',
  sd2: false,
  tanggal_sd2: null,
  tempat_sd2: '',
  sd3: false,
  tanggal_sd3: null,
  tempat_sd3: '',
  kelas_umum: '',
  kelas_khusus: '',
  status_umat: '',
  tim_kerja: '',
  posisi: '',
  ikrar_1: false,
  ikrar_2: false,
  ikrar_3: false,
  ikrar_4: false,
  ikrar_5: false,
  ikrar_6: false,
  ren_chai_pan: false,
  tanggal_ren_chai_pan: null,
  lien_ciang_pan: false,
  tanggal_lien_ciang_pan: null,
  ciang_yen_pan: false,
  tanggal_ciang_yen_pan: null,
  ching_khou: false,
  tanggal_ching_khou: null,
  meninggal: false,
  tanggal_meninggal: null,
  keterangan: '',
  // status: true
})

watch(
  () => formData.value.sd3,
  (val) => {
    if (!val) {
      formData.value.tanggal_sd3 = null
      formData.value.tempat_sd3 = ''
    }
  }
)

watch(
  () => formData.value.sd2,
  (val) => {
    if (!val) {
      formData.value.tanggal_sd2 = null
      formData.value.tempat_sd2 = ''
    }
  }
)

const hasAnCuo = computed({
  get: () => !!formData.value.tanggal_ancuo,
  set: (val: boolean) => {
    if (val) {
      if (!formData.value.tanggal_ancuo) {
        formData.value.tanggal_ancuo = dayjs().format('YYYY-MM-DD')
      }
    } else {
      formData.value.tanggal_ancuo = undefined
      formData.value.nama_cetya_rumah = ''
    }
  }
})

function formatUmatLabel(item: Partial<Umat>): string {
  if (!item) return ''
  const parts: string[] = []
  if (item.nama_indonesia) parts.push(item.nama_indonesia)
  if (item.alias) parts.push(item.alias)
  if (item.nama_mandarin) parts.push(item.nama_mandarin)
  const formatted = parts.join(' / ')
  return formatted ? `${formatted} (${item.kode_buku || item.id || ''})` : item.kode_buku || String(item.id || '')
}

const TRIM_FIELDS: (keyof Umat)[] = [
  'kelas_khusus',
  'kelas_umum',
  'fotang_chiutao',
  'fotang_aktif',
  'tcs',
  'waktu_chiutao_mandarin',
  'pendidikan',
  'pekerjaan',
  'jenis_kelamin',
  'tempat_sd2',
  'tempat_sd3',
  'tim_kerja',
  'posisi',
  'status_umat',
  'wilayah'
]

function trimTargetFields(data: Partial<Umat>): Partial<Umat> {
  const trimmed: Record<string, unknown> = { ...data }
  for (const field of TRIM_FIELDS) {
    if (typeof trimmed[field] === 'string') {
      trimmed[field] = (trimmed[field] as string).trim()
    }
    if (trimmed[field] === '0') {
      trimmed[field] = null
    }
  }
  return trimmed as Partial<Umat>
}

// Watch props for initial data updates on edit page
watch(
  () => props.initialData,
  (val) => {
    if (val && Object.keys(val).length > 0) {
      const trimmedVal = trimTargetFields(val)
      formData.value = { ...formData.value, ...trimmedVal }
    }
  },
  { immediate: true, deep: true }
)

// Form Validation Rules
const rules: FormRules = {
  kode_buku: [
    { required: false, message: 'Kode buku wajib diisi', trigger: 'blur' },
    { min: 2, message: 'Kode Buku minimal 2 karakter', trigger: 'blur' }
  ],
  nama_indonesia: [
    { required: true, message: 'Nama Ciu Tao wajib diisi', trigger: 'blur' }
  ],
  nama_cetya_rumah: [
    {
      validator: (_rule: any, value: any, callback: any) => {
        if (hasAnCuo.value && (!value || !String(value).trim())) {
          callback(new Error('Nama Cetya Rumah wajib diisi jika An Cuo diisi'))
        } else {
          callback()
        }
      },
      trigger: ['blur', 'change']
    }
  ],
  jenis_kelamin: [
    { required: true, message: 'Jenis kelamin wajib diisi', trigger: 'blur' }
  ],
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      const payload = trimTargetFields(formData.value)
      emit('submit', payload)
    }
  })
}

function handleCancel() {
  emit('cancel')
}
</script>

<style scoped>
.umat-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-section-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--el-text-color-primary);
}

.w-full {
  width: 100%;
}

.mb-4 {
  margin-bottom: 1rem;
}

.mt-2 {
  margin-top: 0.5rem;
}

:deep(.el-form-item__label .el-checkbox) {
  margin-right: 0;
  height: auto;
  font-weight: 600;
}

.ikrar-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  padding: 0.25rem 0;
}

.status-date-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}
</style>
