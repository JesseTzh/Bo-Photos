export interface AuthState {
  initialized: boolean;
  authenticated: boolean;
}

export interface ErrorEnvelope {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
}

export interface SuccessEnvelope<T> {
  data: T;
}

export interface PasswordRequest {
  password: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}
