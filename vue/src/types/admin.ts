export interface AdminGroup {
  group_id: number
  group_name: string
  r_insert?: boolean
  r_edit?: boolean
  r_delete?: boolean
  r_reporting?: boolean
  r_position_id?: number
  group_desc?: string
}

export interface Admin {
  id?: number
  username: string
  password?: string
  group_id: number
  email?: string
  phone_number?: string
  img_url?: string
  flag_use: boolean
  date_start?: string
  date_end?: string
  login_desc?: string
  last_login?: string
  department_id?: number
  is_warehouse: boolean
  admin_group?: AdminGroup
}

export interface AdminQueryParams {
  page?: number
  limit?: number
  username?: string
  group_name?: string
}

export interface AdminListResponse {
  data: Admin[]
  meta: {
    page: number
    limit: number
    total: number
  }
  resource?: string
}

export interface AdminSingleResponse {
  data: Admin
  resource?: string
}

export interface AdminGroupListResponse {
  data: AdminGroup[]
  meta?: {
    page: number
    limit: number
    total: number
  }
  resource?: string
}

export interface AdminDepartment {
  department_id: number
  department_code: string
  department_name: string
  status?: boolean
  mod_act?: string
  mod_by?: number
  mod_date?: string
}

export interface AdminDepartmentListResponse {
  data: AdminDepartment[]
  resource?: string
}