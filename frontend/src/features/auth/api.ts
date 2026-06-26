import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";
import type {
  AuthState,
  ChangePasswordRequest,
  PasswordRequest
} from "../../api/schema";

const authStateKey = ["auth", "state"] as const;

export function useAuthState() {
  return useQuery({
    queryKey: authStateKey,
    queryFn: () => apiRequest<AuthState>("/auth/session")
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PasswordRequest) =>
      apiRequest<void>("/auth/login", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: authStateKey })
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiRequest<void>("/auth/logout", { method: "POST" }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: authStateKey })
  });
}

export function useChangePassword() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ChangePasswordRequest) =>
      apiRequest<void>("/admin/password", {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: authStateKey })
  });
}
