export interface EditableExifSummary {
  camera?: string;
  lens?: string;
  exposure_time?: string;
  aperture?: string;
  iso?: string;
  focal_length?: string;
  shoot_at?: string;
  latitude?: number;
  longitude?: number;
  exif_json?: string;
}

function text(value: unknown) {
  if (!value) return undefined;
  if (typeof value === "string" || typeof value === "number") return String(value);
  if (typeof value === "object" && "description" in value) {
    return String((value as { description?: unknown }).description ?? "") || undefined;
  }
  return undefined;
}

export async function readReferenceExif(file: File): Promise<EditableExifSummary> {
  const ExifReader = await import("exifreader");
  const tags = await ExifReader.load(await file.arrayBuffer(), { expanded: true });
  const exif = (tags.exif ?? {}) as Record<string, unknown>;
  const gps = (tags.gps ?? {}) as Record<string, unknown>;
  return {
    camera: text(exif.Model) ?? text(exif.Make),
    lens: text(exif.LensModel),
    exposure_time: text(exif.ExposureTime),
    aperture: text(exif.FNumber),
    iso: text(exif.ISOSpeedRatings),
    focal_length: text(exif.FocalLength),
    shoot_at: normalizeExifDate(text(exif.DateTimeOriginal)),
    latitude: typeof gps.Latitude === "number" ? gps.Latitude : undefined,
    longitude: typeof gps.Longitude === "number" ? gps.Longitude : undefined,
    exif_json: JSON.stringify(tags)
  };
}

export function normalizeExifDate(value?: string) {
  if (!value) return undefined;
  const normalized = value.replace(/^(\d{4}):(\d{2}):(\d{2})/, "$1-$2-$3");
  const date = new Date(normalized);
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}
