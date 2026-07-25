const fs = require('fs');
const path = require('path');

/**
 * Entity Configuration for Types, API, and Views generator.
 * Modeled strictly after umat.ts (Types) and api/umat.ts (API Client).
 */
const entities = [
  {
    key: 'adminGroup',
    className: 'AdminGroup',
    path: 'user-management/AdminGroupView.vue',
    title: 'Master Data Admin Group',
    entity: 'Admin Group',
    endpoint: '/v1/admin-groups',
    pk: 'group_id',
    pkType: 'number',
    fields: [
      { name: 'group_id', type: 'number', optional: false },
      { name: 'group_name', type: 'string', optional: false },
      { name: 'r_insert', type: 'boolean', optional: true },
      { name: 'r_edit', type: 'boolean', optional: true },
      { name: 'r_delete', type: 'boolean', optional: true },
      { name: 'r_reporting', type: 'boolean', optional: true },
      { name: 'r_position_id', type: 'number', optional: true },
      { name: 'group_desc', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'group_name', type: 'string' },
      { name: 'group_desc', type: 'string' }
    ],
    columns: [
      { prop: 'group_id', label: 'ID', width: '80', align: 'center', sortable: true },
      { prop: 'group_name', label: 'Nama Group', minWidth: '180', bold: true },
      { prop: 'group_desc', label: 'Keterangan', minWidth: '220' },
      { prop: 'r_insert', label: 'Hak Tambah', width: '110', type: 'boolean' },
      { prop: 'r_edit', label: 'Hak Edit', width: '100', type: 'boolean' },
      { prop: 'r_delete', label: 'Hak Hapus', width: '100', type: 'boolean' },
      { prop: 'r_reporting', label: 'Hak Laporan', width: '120', type: 'boolean' }
    ],
    filters: [
      { prop: 'group_name', label: 'Nama Group', placeholder: 'Cari nama group...' },
      { prop: 'group_desc', label: 'Keterangan', placeholder: 'Cari deskripsi...' }
    ]
  },
  {
    key: 'groupMenuMapping',
    className: 'GroupMenuMapping',
    path: 'user-management/GroupMenuMappingView.vue',
    title: 'Group Menu Mapping',
    entity: 'Group Menu Mapping',
    endpoint: '/v1/group-menu-mappings',
    pk: 'menu_id',
    pkType: 'number',
    fields: [
      { name: 'menu_id', type: 'number', optional: false },
      { name: 'parent_id', type: 'number', optional: true },
      { name: 'menu_name', type: 'string', optional: false },
      { name: 'page_url', type: 'string', optional: true },
      { name: 'sequence', type: 'number', optional: true },
      { name: 'menu_desc', type: 'string', optional: true },
      { name: 'parent_level_1', type: 'number', optional: true },
      { name: 'flag_active', type: 'boolean', optional: true }
    ],
    queryParams: [
      { name: 'menu_name', type: 'string' },
      { name: 'page_url', type: 'string' }
    ],
    columns: [
      { prop: 'menu_id', label: 'ID', width: '80', align: 'center', sortable: true },
      { prop: 'menu_name', label: 'Nama Menu', minWidth: '180', bold: true },
      { prop: 'page_url', label: 'URL Halaman', minWidth: '200' },
      { prop: 'sequence', label: 'Urutan', width: '90', align: 'center' },
      { prop: 'menu_desc', label: 'Keterangan', minWidth: '200' },
      { prop: 'flag_active', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'menu_name', label: 'Nama Menu', placeholder: 'Cari nama menu...' },
      { prop: 'page_url', label: 'URL Halaman', placeholder: 'Cari page url...' }
    ]
  },
  {
    key: 'adminSubWarehouse',
    className: 'AdminSubWarehouse',
    path: 'user-management/AdminSubWarehouseView.vue',
    title: 'Admin Sub Warehouse',
    entity: 'Admin Sub Warehouse',
    endpoint: '/v1/admin-sub-warehouses',
    pk: 'sub_wh_id',
    pkType: 'number',
    fields: [
      { name: 'sub_wh_id', type: 'number', optional: false },
      { name: 'wh_id', type: 'number', optional: true },
      { name: 'full_name', type: 'string', optional: false },
      { name: 'pic', type: 'string', optional: true },
      { name: 'doc_code', type: 'string', optional: true },
      { name: 'cru_id', type: 'number', optional: true },
      { name: 'update_uid', type: 'number', optional: true },
      { name: 'lst_update', type: 'string', optional: true },
      { name: 'flag_productions', type: 'boolean', optional: true },
      { name: 'sub_wh_type', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'full_name', type: 'string' },
      { name: 'pic', type: 'string' },
      { name: 'doc_code', type: 'string' },
      { name: 'sub_wh_type', type: 'string' }
    ],
    columns: [
      { prop: 'sub_wh_id', label: 'ID', width: '80', align: 'center', sortable: true },
      { prop: 'full_name', label: 'Nama Sub Warehouse', minWidth: '200', bold: true },
      { prop: 'pic', label: 'PIC', minWidth: '150' },
      { prop: 'doc_code', label: 'Kode Dokumen', minWidth: '140' },
      { prop: 'sub_wh_type', label: 'Tipe Sub WH', minWidth: '130' },
      { prop: 'flag_productions', label: 'Produksi', width: '110', type: 'boolean' }
    ],
    filters: [
      { prop: 'full_name', label: 'Nama Warehouse', placeholder: 'Cari nama sub wh...' },
      { prop: 'pic', label: 'PIC', placeholder: 'Cari PIC...' },
      { prop: 'doc_code', label: 'Kode Dokumen', placeholder: 'Cari doc code...' }
    ]
  },
  {
    key: 'topic',
    className: 'Topic',
    path: 'master-data/TopicView.vue',
    title: 'Master Data Topic',
    entity: 'Topic',
    endpoint: '/v1/topics',
    pk: 'topic_code',
    pkType: 'string',
    fields: [
      { name: 'topic_code', type: 'string', optional: false },
      { name: 'topic_name', type: 'string', optional: false },
      { name: 'topic_category', type: 'string', optional: true },
      { name: 'description', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'mod_act', type: 'string', optional: true },
      { name: 'mod_by', type: 'string', optional: true },
      { name: 'mod_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'topic_code', type: 'string' },
      { name: 'topic_name', type: 'string' },
      { name: 'topic_category', type: 'string' }
    ],
    columns: [
      { prop: 'topic_code', label: 'Kode Topik', width: '130', align: 'center', tag: true },
      { prop: 'topic_name', label: 'Nama Topik', minWidth: '200', bold: true },
      { prop: 'topic_category', label: 'Kategori', minWidth: '150' },
      { prop: 'description', label: 'Keterangan', minWidth: '220' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'topic_code', label: 'Kode Topik', placeholder: 'Cari kode topik...' },
      { prop: 'topic_name', label: 'Nama Topik', placeholder: 'Cari nama topik...' },
      { prop: 'topic_category', label: 'Kategori', placeholder: 'Cari kategori...' }
    ]
  },
  {
    key: 'activity',
    className: 'Activity',
    path: 'master-data/ActivityView.vue',
    title: 'Master Data Activity',
    entity: 'Activity',
    endpoint: '/v1/activities',
    pk: 'event_code',
    pkType: 'string',
    fields: [
      { name: 'event_code', type: 'string', optional: false },
      { name: 'event_name', type: 'string', optional: false },
      { name: 'event_category', type: 'string', optional: true },
      { name: 'description', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'mod_act', type: 'string', optional: true },
      { name: 'mod_by', type: 'string', optional: true },
      { name: 'mod_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'event_code', type: 'string' },
      { name: 'event_name', type: 'string' },
      { name: 'event_category', type: 'string' }
    ],
    columns: [
      { prop: 'event_code', label: 'Kode Event', width: '130', align: 'center', tag: true },
      { prop: 'event_name', label: 'Nama Event', minWidth: '200', bold: true },
      { prop: 'event_category', label: 'Kategori', minWidth: '150' },
      { prop: 'description', label: 'Keterangan', minWidth: '220' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'event_code', label: 'Kode Event', placeholder: 'Cari kode event...' },
      { prop: 'event_name', label: 'Nama Event', placeholder: 'Cari nama event...' },
      { prop: 'event_category', label: 'Kategori', placeholder: 'Cari kategori...' }
    ]
  },
  {
    key: 'timKerja',
    className: 'TimKerja',
    path: 'master-data/TimKerjaView.vue',
    title: 'Master Data Tim Kerja',
    entity: 'Tim Kerja',
    endpoint: '/v1/tim-kerja',
    pk: 'lookup_id',
    pkType: 'string',
    fields: [
      { name: 'lookup_id', type: 'string', optional: false },
      { name: 'category_id', type: 'string', optional: true },
      { name: 'lookup_value', type: 'string', optional: false },
      { name: 'lookup_description', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'mod_act', type: 'string', optional: true },
      { name: 'mod_by', type: 'string', optional: true },
      { name: 'mod_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'lookup_id', type: 'string' },
      { name: 'lookup_value', type: 'string' },
      { name: 'lookup_description', type: 'string' }
    ],
    columns: [
      { prop: 'lookup_id', label: 'Lookup ID', width: '130', align: 'center', tag: true },
      { prop: 'lookup_value', label: 'Posisi / Tim Kerja', minWidth: '200', bold: true },
      { prop: 'lookup_description', label: 'Keterangan', minWidth: '220' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'lookup_id', label: 'Lookup ID', placeholder: 'Cari lookup id...' },
      { prop: 'lookup_value', label: 'Posisi / Tim', placeholder: 'Cari posisi/tim...' },
      { prop: 'lookup_description', label: 'Keterangan', placeholder: 'Cari deskripsi...' }
    ]
  },
  {
    key: 'tahunCiuTao',
    className: 'TahunCiuTao',
    path: 'master-data/TahunCiuTaoView.vue',
    title: 'Master Data Tahun Ciu Tao',
    entity: 'Tahun Ciu Tao',
    endpoint: '/v1/tahun-ciu-tao',
    pk: 'tahun_mandarin',
    pkType: 'string',
    fields: [
      { name: 'tahun_mandarin', type: 'string', optional: false },
      { name: 'start_date', type: 'string', optional: true },
      { name: 'end_date', type: 'string', optional: true },
      { name: 'description', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'mod_act', type: 'string', optional: true },
      { name: 'mod_by', type: 'string', optional: true },
      { name: 'mod_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'tahun_mandarin', type: 'string' },
      { name: 'description', type: 'string' }
    ],
    columns: [
      { prop: 'tahun_mandarin', label: 'Tahun Mandarin', minWidth: '160', bold: true },
      { prop: 'start_date', label: 'Tanggal Mulai', minWidth: '140', type: 'date' },
      { prop: 'end_date', label: 'Tanggal Selesai', minWidth: '140', type: 'date' },
      { prop: 'description', label: 'Keterangan', minWidth: '220' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'tahun_mandarin', label: 'Tahun Mandarin', placeholder: 'Cari tahun (e.g. 2024)...' },
      { prop: 'description', label: 'Keterangan', placeholder: 'Cari deskripsi...' }
    ]
  },
  {
    key: 'penggalangDana',
    className: 'PenggalangDana',
    path: 'master-data/PenggalangDanaView.vue',
    title: 'Master Data Penggalang Dana',
    entity: 'Penggalang Dana',
    endpoint: '/v1/penggalang-dana',
    pk: 'id',
    pkType: 'number',
    fields: [
      { name: 'id', type: 'number', optional: false },
      { name: 'no', type: 'string', optional: true },
      { name: 'nama', type: 'string', optional: false },
      { name: 'mandarin', type: 'string', optional: true },
      { name: 'keterangan', type: 'string', optional: true },
      { name: 'lookup_fothang', type: 'number', optional: true },
      { name: 'alamat', type: 'string', optional: true },
      { name: 'telepon', type: 'string', optional: true },
      { name: 'mobile', type: 'string', optional: true },
      { name: 'email', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'created_by', type: 'number', optional: true },
      { name: 'created_date', type: 'string', optional: true },
      { name: 'updated_by', type: 'number', optional: true },
      { name: 'updated_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'nama', type: 'string' },
      { name: 'mandarin', type: 'string' },
      { name: 'fotang', type: 'number' }
    ],
    columns: [
      { prop: 'id', label: 'ID', width: '80', align: 'center', sortable: true },
      { prop: 'no', label: 'No', width: '120', tag: true },
      { prop: 'nama', label: 'Nama', minWidth: '180', bold: true },
      { prop: 'mandarin', label: 'Nama Mandarin', minWidth: '140' },
      { prop: 'mobile', label: 'No HP / Telepon', minWidth: '140' },
      { prop: 'keterangan', label: 'Keterangan', minWidth: '200' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'nama', label: 'Nama', placeholder: 'Cari nama...' },
      { prop: 'mandarin', label: 'Mandarin', placeholder: 'Cari nama mandarin...' }
    ]
  },
  {
    key: 'sxyDonatur',
    className: 'SxyDonatur',
    path: 'master-data/SxyDonaturView.vue',
    title: 'Master Data Sxy Donatur',
    entity: 'Sxy Donatur',
    endpoint: '/v1/sxy-donatur',
    pk: 'id',
    pkType: 'number',
    fields: [
      { name: 'id', type: 'number', optional: false },
      { name: 'no', type: 'string', optional: true },
      { name: 'nama', type: 'string', optional: false },
      { name: 'mandarin', type: 'string', optional: true },
      { name: 'keterangan', type: 'string', optional: true },
      { name: 'lookup_fothang', type: 'number', optional: true },
      { name: 'alamat', type: 'string', optional: true },
      { name: 'telepon', type: 'string', optional: true },
      { name: 'mobile', type: 'string', optional: true },
      { name: 'email', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'created_by', type: 'number', optional: true },
      { name: 'created_date', type: 'string', optional: true },
      { name: 'updated_by', type: 'number', optional: true },
      { name: 'updated_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'nama', type: 'string' },
      { name: 'mandarin', type: 'string' },
      { name: 'fotang', type: 'number' }
    ],
    columns: [
      { prop: 'id', label: 'ID', width: '80', align: 'center', sortable: true },
      { prop: 'no', label: 'No Donatur', width: '130', tag: true },
      { prop: 'nama', label: 'Nama Donatur', minWidth: '180', bold: true },
      { prop: 'mandarin', label: 'Nama Mandarin', minWidth: '140' },
      { prop: 'mobile', label: 'No HP / Telepon', minWidth: '140' },
      { prop: 'keterangan', label: 'Keterangan', minWidth: '200' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'nama', label: 'Nama', placeholder: 'Cari nama donatur...' },
      { prop: 'mandarin', label: 'Mandarin', placeholder: 'Cari nama mandarin...' }
    ]
  },
  {
    key: 'kelas',
    className: 'Kelas',
    path: 'transaction/KelasView.vue',
    title: 'Transaksi Kelas',
    entity: 'Kelas',
    endpoint: '/v1/kelas',
    pk: 'lookup_id',
    pkType: 'string',
    fields: [
      { name: 'lookup_id', type: 'string', optional: false },
      { name: 'category_id', type: 'string', optional: true },
      { name: 'lookup_value', type: 'string', optional: false },
      { name: 'lookup_description', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'mod_act', type: 'string', optional: true },
      { name: 'mod_by', type: 'string', optional: true },
      { name: 'mod_date', type: 'string', optional: true }
    ],
    queryParams: [
      { name: 'lookup_id', type: 'string' },
      { name: 'lookup_value', type: 'string' },
      { name: 'lookup_description', type: 'string' }
    ],
    columns: [
      { prop: 'lookup_id', label: 'Kode Kelas', width: '130', align: 'center', tag: true },
      { prop: 'lookup_value', label: 'Nama Kelas', minWidth: '200', bold: true },
      { prop: 'lookup_description', label: 'Keterangan', minWidth: '220' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'lookup_id', label: 'Kode Kelas', placeholder: 'Cari kode kelas...' },
      { prop: 'lookup_value', label: 'Nama Kelas', placeholder: 'Cari nama kelas...' }
    ]
  },
  {
    key: 'donasiSxy',
    className: 'DonasiSxy',
    path: 'transaction/DonasiSxyView.vue',
    title: 'Transaksi Donasi Sxy',
    entity: 'Donasi Sxy',
    endpoint: '/v1/donasi-sxy',
    pk: 'id',
    pkType: 'number',
    fields: [
      { name: 'id', type: 'number', optional: false },
      { name: 'no_kwitansi', type: 'string', optional: false },
      { name: 'tanggal', type: 'string', optional: true },
      { name: 'donatur_id', type: 'number', optional: true },
      { name: 'penggalang_id', type: 'number', optional: true },
      { name: 'jumlah', type: 'number', optional: true },
      { name: 'tipe_sumbangan', type: 'number', optional: true },
      { name: 'no_kupon', type: 'string', optional: true },
      { name: 'keterangan', type: 'string', optional: true },
      { name: 'status', type: 'boolean', optional: true },
      { name: 'created_by', type: 'number', optional: true },
      { name: 'created_date', type: 'string', optional: true },
      { name: 'updated_by', type: 'number', optional: true },
      { name: 'updated_date', type: 'string', optional: true },
      { name: 'tanggal_transfer', type: 'string', optional: true },
      { name: 'atas_nama', type: 'string', optional: true },
      { name: 'ttk_sent', type: 'boolean', optional: true }
    ],
    queryParams: [
      { name: 'no_kwitansi', type: 'string' },
      { name: 'donatur_id', type: 'number' },
      { name: 'tanggal', type: 'string' }
    ],
    columns: [
      { prop: 'id', label: 'ID', width: '80', align: 'center', sortable: true },
      { prop: 'no_kwitansi', label: 'No. Kwitansi', minWidth: '160', bold: true, tag: true },
      { prop: 'tanggal', label: 'Tanggal', minWidth: '140', type: 'date' },
      { prop: 'jumlah', label: 'Jumlah (Rp)', minWidth: '150', type: 'currency' },
      { prop: 'no_kupon', label: 'No. Kupon', minWidth: '140' },
      { prop: 'keterangan', label: 'Keterangan', minWidth: '200' },
      { prop: 'status', label: 'Status', width: '110', type: 'statusBool' }
    ],
    filters: [
      { prop: 'no_kwitansi', label: 'No. Kwitansi', placeholder: 'Cari no kwitansi...' }
    ]
  }
];

// Helper to ensure target output directory exists
function ensureDirExists(filePath) {
  const dir = path.dirname(filePath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

// 1. GENERATE TYPES FILES (`src/types/<key>.ts`)
entities.forEach(item => {
  const typeFilePath = path.join(__dirname, 'src', 'types', `${item.key}.ts`);
  ensureDirExists(typeFilePath);

  const fieldLines = item.fields.map(f => `  ${f.name}${f.optional ? '?' : ''}: ${f.type}`).join('\n');
  const queryParamLines = item.queryParams.map(q => `  ${q.name}?: ${q.type}`).join('\n');

  const typeContent = `export interface ${item.className} {
${fieldLines}
}

export interface ${item.className}QueryParams {
  page?: number
  limit?: number
${queryParamLines}
}

export interface PaginatedMeta {
  page: number
  limit: number
  total: number
}

export interface ${item.className}ListResponse {
  data: ${item.className}[]
  meta: PaginatedMeta
  resource?: string
}

export interface ${item.className}SingleResponse {
  data: ${item.className}
  resource?: string
}
`;

  fs.writeFileSync(typeFilePath, typeContent);
  console.log(`Generated Type: src/types/${item.key}.ts`);
});

// 2. GENERATE API CLIENT FILES (`src/api/<key>.ts`)
entities.forEach(item => {
  const apiFilePath = path.join(__dirname, 'src', 'api', `${item.key}.ts`);
  ensureDirExists(apiFilePath);

  const cleanParamBindings = item.queryParams.map(q => `    if (params.${q.name}) cleanParams.${q.name} = params.${q.name}`).join('\n');

  const apiContent = `import apiClient from './client'
import type {
  ${item.className},
  ${item.className}QueryParams,
  ${item.className}ListResponse,
  ${item.className}SingleResponse
} from '../types/${item.key}'

export const ${item.key}Api = {
  /**
   * Fetch paginated list of ${item.entity} with optional search filters.
   */
  async get${item.className}s(
    params: ${item.className}QueryParams = {},
    signal?: AbortSignal
  ): Promise<${item.className}ListResponse> {
    const cleanParams: Record<string, any> = {
      page: params.page || 1,
      limit: params.limit || 10
    }

${cleanParamBindings}

    const response = await apiClient.get<${item.className}ListResponse>('${item.endpoint}', {
      params: cleanParams,
      signal
    })
    return response.data
  },

  /**
   * Fetch single ${item.className} details by ID.
   */
  async get${item.className}ById(
    id: ${item.pkType},
    signal?: AbortSignal
  ): Promise<${item.className}SingleResponse> {
    const response = await apiClient.get<${item.className}SingleResponse>(\`${item.endpoint}/\${id}\`, { signal })
    return response.data
  },

  /**
   * Create a new ${item.className} record.
   */
  async create${item.className}(payload: Partial<${item.className}>): Promise<${item.className}SingleResponse> {
    const response = await apiClient.post<${item.className}SingleResponse>('${item.endpoint}', payload)
    return response.data
  },

  /**
   * Update an existing ${item.className} record.
   */
  async update${item.className}(
    id: ${item.pkType},
    payload: Partial<${item.className}>
  ): Promise<${item.className}SingleResponse> {
    const response = await apiClient.patch<${item.className}SingleResponse>(\`${item.endpoint}/\${id}\`, payload)
    return response.data
  },

  /**
   * Delete a ${item.className} record by ID.
   */
  async delete${item.className}(id: ${item.pkType}): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(\`${item.endpoint}/\${id}\`)
    return response.data
  }
}

export default ${item.key}Api
`;

  fs.writeFileSync(apiFilePath, apiContent);
  console.log(`Generated API: src/api/${item.key}.ts`);
});

// 3. GENERATE VUE VIEWS (`src/views/<path>`)
entities.forEach(item => {
  const viewFilePath = path.join(__dirname, 'src', 'views', item.path);
  ensureDirExists(viewFilePath);

  // Generate Filter inputs markup
  const filterInputsMarkup = item.filters.map(f => `        <el-col :xs="24" :sm="12" :md="6">
          <el-form-item label="${f.label}">
            <el-input
              v-model="filters.${f.prop}"
              placeholder="${f.placeholder}"
              clearable
              :prefix-icon="Search"
              @input="onFilterChange"
            />
          </el-form-item>
        </el-col>`).join('\n\n');

  // Generate Table columns markup
  const tableColumnsMarkup = item.columns.map(col => {
    if (col.type === 'statusBool') {
      return `        <el-table-column prop="${col.prop}" label="${col.label}" width="${col.width || 110}" align="${col.align || 'center'}">
          <template #default="{ row }">
            <el-tag :type="row.${col.prop} ? 'success' : 'info'" size="small">
              {{ row.${col.prop} ? 'Aktif' : 'Nonaktif' }}
            </el-tag>
          </template>
        </el-table-column>`;
    }
    if (col.type === 'boolean') {
      return `        <el-table-column prop="${col.prop}" label="${col.label}" width="${col.width || 110}" align="${col.align || 'center'}">
          <template #default="{ row }">
            <el-tag :type="row.${col.prop} ? 'primary' : 'info'" size="small" effect="plain">
              {{ row.${col.prop} ? 'Ya' : 'Tidak' }}
            </el-tag>
          </template>
        </el-table-column>`;
    }
    if (col.type === 'currency') {
      return `        <el-table-column prop="${col.prop}" label="${col.label}" min-width="${col.minWidth || 150}">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.${col.prop} ? 'Rp ' + Number(row.${col.prop}).toLocaleString('id-ID') : '-' }}</span>
          </template>
        </el-table-column>`;
    }
    if (col.tag) {
      return `        <el-table-column prop="${col.prop}" label="${col.label}" width="${col.width || 130}" align="${col.align || 'left'}">
          <template #default="{ row }">
            <el-tag size="small" type="info" class="font-mono">{{ row.${col.prop} }}</el-tag>
          </template>
        </el-table-column>`;
    }
    if (col.bold) {
      return `        <el-table-column prop="${col.prop}" label="${col.label}" min-width="${col.minWidth || 180}">
          <template #default="{ row }">
            <span class="font-semibold">{{ row.${col.prop} || '-' }}</span>
          </template>
        </el-table-column>`;
    }
    return `        <el-table-column prop="${col.prop}" label="${col.label}" ${col.width ? `width="${col.width}"` : ''} ${col.minWidth ? `min-width="${col.minWidth}"` : ''} ${col.align ? `align="${col.align}"` : ''} ${col.sortable ? 'sortable' : ''}>
          <template #default="{ row }">
            <span>{{ row.${col.prop} || '-' }}</span>
          </template>
        </el-table-column>`;
  }).join('\n\n');

  // Filter reactive object properties
  const filterPropsInit = item.filters.map(f => `  ${f.prop}: ''`).join(',\n');

  // Clean params assignment in fetch function
  const filterParamsFetch = item.filters.map(f => `        ${f.prop}: filters.${f.prop}?.trim()`).join(',\n');

  // Reset filters reset assignments
  const filterResetAssignments = item.filters.map(f => `  filters.${f.prop} = ''`).join('\n');

  const viewContent = `<template>
  <div class="view-container">
    <!-- Header Section -->
    <div class="page-header">
      <div>
        <h2 class="page-title">${item.title}</h2>
        <p class="page-subtitle">Kelola daftar data ${item.entity.toLowerCase()}, pencarian, serta pembaruan profil</p>
      </div>
      <el-button
        type="primary"
        size="large"
        :icon="Plus"
        class="create-btn"
        @click="handleCreate"
      >
        Tambah ${item.entity} Baru
      </el-button>
    </div>

    <!-- Filter Card -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-header">
        <el-icon class="filter-icon"><Search /></el-icon>
        <span class="filter-title">Filter & Pencarian Data</span>
      </div>

      <el-row :gutter="16" class="filter-row">
${filterInputsMarkup}
      </el-row>

      <div class="filter-actions">
        <el-button :icon="Refresh" @click="resetFilters">Reset Filter</el-button>
      </div>
    </el-card>

    <!-- Table Card -->
    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="dataList"
        stripe
        border
        height="475"
        style="width: 100%"
        empty-text="Tidak ada data ${item.entity.toLowerCase()} yang ditemukan"
      >
        <el-table-column :index="getRowIndex" type="index" label="No." width="70" align="center" fixed="left" />

${tableColumnsMarkup}

        <!-- Actions Column -->
        <el-table-column label="Aksi" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="primary"
                size="small"
                circle
                :icon="Edit"
                title="Edit ${item.entity}"
                @click="handleEdit(row.${item.pk})"
              />

              <el-popconfirm
                title="Apakah Anda yakin ingin menghapus data ini?"
                confirm-button-text="Ya, Hapus"
                cancel-button-text="Batal"
                confirm-button-type="danger"
                @confirm="handleDelete(row.${item.pk})"
              >
                <template #reference>
                  <el-button
                    type="danger"
                    size="small"
                    circle
                    :icon="Delete"
                    title="Hapus ${item.entity}"
                  />
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import { ${item.key}Api } from '../../api/${item.key}'
import type { ${item.className}, ${item.className}QueryParams } from '../../types/${item.key}'

const router = useRouter()

// Memory Optimization: shallowRef for table dataset
const dataList = shallowRef<${item.className}[]>([])
const loading = ref(false)

// Pagination state
const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0
})

const getRowIndex = (index: number) => {
  return (pagination.page - 1) * pagination.limit + index + 1
}

// Search Filter state
const filters = reactive<${item.className}QueryParams>({
${filterPropsInit}
})

let currentAbortController: AbortController | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

async function fetchData() {
  if (currentAbortController) {
    currentAbortController.abort()
  }
  currentAbortController = new AbortController()
  loading.value = true

  try {
    const res = await ${item.key}Api.get${item.className}s(
      {
        page: pagination.page,
        limit: pagination.limit,
${filterParamsFetch}
      },
      currentAbortController.signal
    )

    dataList.value = res.data || []
    pagination.total = res.meta?.total || 0
  } catch (err: any) {
    if (err.name === 'CanceledError' || err.name === 'AbortError') return
    console.error('Error fetching data:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal memuat data')
  } finally {
    loading.value = false
  }
}

function onFilterChange() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    pagination.page = 1
    fetchData()
  }, 300)
}

function resetFilters() {
${filterResetAssignments}
  pagination.page = 1
  fetchData()
}

function handleSizeChange(newLimit: number) {
  pagination.limit = newLimit
  pagination.page = 1
  fetchData()
}

function handlePageChange(newPage: number) {
  pagination.page = newPage
  fetchData()
}

function handleCreate() {
  ElNotification({
    title: 'Informasi',
    message: 'Tambah ${item.entity.toLowerCase()} baru',
    type: 'info'
  })
}

function handleEdit(id: ${item.pkType}) {
  ElNotification({
    title: 'Informasi',
    message: \`Edit ${item.entity.toLowerCase()} ID/Kode: \${id}\`,
    type: 'info'
  })
}

async function handleDelete(id: ${item.pkType}) {
  try {
    await ${item.key}Api.delete${item.className}(id)
    ElNotification({
      title: 'Berhasil',
      message: \`Data ${item.entity.toLowerCase()} \${id} berhasil dihapus\`,
      type: 'success'
    })
    fetchData()
  } catch (err: any) {
    console.error('Failed to delete item:', err)
    ElMessage.error(err.response?.data?.error || err.message || 'Gagal menghapus data')
  }
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  if (currentAbortController) currentAbortController.abort()
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin: 0;
}

.page-subtitle {
  font-size: 0.875rem;
  color: var(--el-text-color-secondary);
  margin: 0.25rem 0 0 0;
}

.create-btn {
  font-weight: 600;
  border-radius: 8px;
}

.filter-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.filter-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.filter-icon {
  color: var(--el-color-primary);
  font-size: 1.1rem;
}

.filter-row {
  margin-bottom: -0.5rem;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

.table-card {
  border-radius: 10px;
  background-color: var(--el-bg-color-overlay);
}

.font-mono {
  font-family: monospace;
}

.font-semibold {
  font-weight: 600;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 1.25rem;
}
</style>
`;

  fs.writeFileSync(viewFilePath, viewContent);
  console.log(`Generated View: src/views/${item.path}`);
});
