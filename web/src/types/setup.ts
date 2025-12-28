// TypeScript types matching Go structs (camelCase JSON)

export interface CheckResult {
  name: string
  passed: boolean
  message?: string
}

export interface ValidationResponse {
  success: boolean
  checks: CheckResult[]
  errorCode?: string
  message?: string
}

export interface APIError {
  code: string
  message: string
  details?: string
}

export interface InitializeResponse {
  success: boolean
  message?: string
  errorCode?: string
}

export interface DatabaseStatus {
  initialized: boolean
  path: string
}
