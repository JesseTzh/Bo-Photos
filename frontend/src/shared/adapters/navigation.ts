import { useLocation, useNavigate, useParams } from "react-router-dom";

export function usePathname() {
  return useLocation().pathname;
}

export function useAppRouter() {
  const navigate = useNavigate();
  return {
    push: (href: string) => navigate(href),
    replace: (href: string) => navigate(href, { replace: true }),
    back: () => window.history.back(),
    prefetch: (_href: string) => undefined
  };
}

export { useParams };
