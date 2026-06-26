import type { ImgHTMLAttributes } from "react";

type AppImageProps = Omit<ImgHTMLAttributes<HTMLImageElement>, "src"> & {
  src?: string | null;
  fill?: boolean;
  priority?: boolean;
  sizes?: string;
};

export function AppImage({ fill, priority, style, loading, src, ...props }: AppImageProps) {
  return (
    <img
      src={src || ""}
      loading={priority ? "eager" : loading}
      style={fill ? { position: "absolute", inset: 0, width: "100%", height: "100%", ...style } : style}
      {...props}
    />
  );
}
