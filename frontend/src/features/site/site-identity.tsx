import { useEffect } from "react";
import { usePublicSettings } from "./api";

export function SiteIdentity() {
  const settings = usePublicSettings().data;

  useEffect(() => {
    document.title = settings?.site_title?.trim() || "BoPhoto";
  }, [settings?.site_title]);

  useEffect(() => {
    const favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    if (!favicon) return;
    favicon.href = settings?.site_favicon_url?.trim() || "/favicon.svg";
  }, [settings?.site_favicon_url]);

  return null;
}
