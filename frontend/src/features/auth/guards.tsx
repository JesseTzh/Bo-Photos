import { Spin } from "antd";
import type { PropsWithChildren } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuthState } from "./api";

function FullPageLoading() {
  return (
    <div className="page-loading" role="status" aria-label="正在加载">
      <Spin size="large" />
    </div>
  );
}

export function LoginRoute({ children }: PropsWithChildren) {
  const auth = useAuthState();
  if (auth.isPending) return <FullPageLoading />;
  if (auth.data?.authenticated) return <Navigate to="/admin" replace />;
  return children;
}

export function AdminRoute({ children }: PropsWithChildren) {
  const auth = useAuthState();
  const location = useLocation();
  if (auth.isPending) return <FullPageLoading />;
  if (!auth.data?.authenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return children;
}

export function PrivateRoute({ children }: PropsWithChildren) {
  const auth = useAuthState();
  const location = useLocation();
  if (auth.isPending) return <FullPageLoading />;
  if (!auth.data?.authenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return children;
}
