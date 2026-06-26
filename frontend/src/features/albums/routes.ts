export function albumPublicHref(value: string, search = "") {
  const normalized = value.startsWith("/") ? value : `/${value}`;
  return `${normalized}${search}`;
}

export function albumGalleryHref(value: string, search = "") {
  return `/gallery/${encodeURIComponent(value)}${search}`;
}
