export interface RegisterPayload {
  username: string
  password: string
}

export type LoginPayload = RegisterPayload

export interface LoginResponse {
  token: string
}
