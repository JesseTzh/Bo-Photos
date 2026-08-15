import { Cross1Icon, HamburgerMenuIcon, MoonIcon, SunIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useAuthState } from "../../features/auth/api";
import { useAlbums, type Album } from "../../features/albums/api";
import { albumPublicHref } from "../../features/albums/routes";
import { usePublicSettings } from "../../features/site/api";
import { AppLink } from "../adapters/link";
import { usePathname } from "../adapters/navigation";
import { useTheme } from "../adapters/theme";
import { useTranslations } from "../adapters/i18n";
import { cn } from "../lib/utils";

const navLinks = [
  { name: "序章", href: "/" },
  { name: "城隅寻迹", href: "/covers" },
  { name: "景行集", href: "/gallery" },
  { name: "关于我", href: "/about" }
];

function albumHref(album: Album) {
  return albumPublicHref(album.album_value);
}

export function UnifiedNav({ hideThemeToggle = false }: { hideThemeToggle?: boolean }) {
  const [isScrolled, setIsScrolled] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const pathname = usePathname();
  const t = useTranslations();
  const albums = useAlbums();
  const settings = usePublicSettings();
  const auth = useAuthState();
  const { resolvedTheme, toggle } = useTheme();
  const navRef = useRef<HTMLElement>(null);

  const siteTitle = settings.data?.site_title?.trim() || "BoPhoto";
  const visibleAlbums = useMemo(
    () => (albums.data?.items ?? []).filter((album) => album.album_value !== "/" && album.visible),
    [albums.data?.items]
  );

  const handleToggle = useCallback(() => {
    toggle();
  }, [toggle]);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 20);
    handleScroll();
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    setMobileMenuOpen(false);
  }, [pathname]);

  const isActive = (path: string) => (path === "/" ? pathname === "/" : pathname.startsWith(path));
  const consoleHref = auth.data?.authenticated ? "/admin" : "/login";
  const consoleLabel = auth.data?.authenticated ? t("Link.dashboard") : t("Login.signIn");
  const showPrivateLink = Boolean(auth.data?.authenticated);

  return (
    <>
      <nav
        ref={navRef}
        className={cn(
          "fixed left-0 top-0 z-50 h-14 w-full transition-all duration-500 ease-out",
          isScrolled || pathname !== "/"
            ? "public-nav-scrolled border-b border-border/30 bg-background/80 backdrop-blur-2xl"
            : "border-b border-transparent bg-transparent"
        )}
      >
        <div className="mx-auto flex h-full max-w-[1400px] items-center justify-between px-5 md:px-8">
          <AppLink href="/" className="group relative flex-shrink-0">
            <span className="font-serif text-[22px] font-medium tracking-[-0.02em] text-foreground transition-opacity duration-300 group-hover:opacity-60">
              {siteTitle}
            </span>
          </AppLink>

          <div className="hidden items-center lg:flex">
            {navLinks.map((link, index) => {
              const active = isActive(link.href);
              return (
                <AppLink
                  key={link.href}
                  href={link.href}
                  onMouseEnter={() => setHoveredIndex(index)}
                  onMouseLeave={() => setHoveredIndex(null)}
                  className={cn(
                    "relative px-4 py-1.5 text-[15px] tracking-[0.02em] transition-all duration-300 ease-out",
                    active ? "font-medium text-foreground" : "font-normal text-muted-foreground hover:text-foreground"
                  )}
                >
                  <span className="relative z-10">{link.name}</span>
                  {active ? (
                    <span className="absolute -bottom-0.5 left-1/2 h-[3px] w-[3px] -translate-x-1/2 rounded-full bg-primary transition-all duration-500" />
                  ) : null}
                  {!active && hoveredIndex === index ? (
                    <span className="absolute inset-0 rounded-full bg-muted/60 transition-all duration-300" />
                  ) : null}
                </AppLink>
              );
            })}
            <span className="mx-3 h-3 w-px bg-border/60" />
            {showPrivateLink ? (
              <AppLink
                href="/private"
                className={cn(
                  "px-4 py-1.5 text-[15px] tracking-[0.02em] transition-all duration-300",
                  isActive("/private")
                    ? "font-medium text-foreground"
                    : "font-normal text-muted-foreground hover:text-foreground"
                )}
              >
                隐私相册
              </AppLink>
            ) : null}
            <AppLink
              href={consoleHref}
              className={cn(
                "px-4 py-1.5 text-[15px] tracking-[0.02em] transition-all duration-300",
                isActive("/admin") || isActive("/login")
                  ? "font-medium text-foreground"
                  : "font-normal text-muted-foreground hover:text-foreground"
              )}
            >
              {consoleLabel}
            </AppLink>
            {mounted && !hideThemeToggle ? (
              <button
                onClick={handleToggle}
                className="ml-2 inline-flex h-9 w-9 items-center justify-center rounded-full transition-all duration-300 hover:bg-muted/60"
                style={{ touchAction: "manipulation" }}
                aria-label={resolvedTheme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
                type="button"
              >
                {resolvedTheme === "dark" ? <SunIcon className="h-4 w-4 text-foreground" /> : <MoonIcon className="h-4 w-4 text-foreground" />}
              </button>
            ) : null}
          </div>

          <div className="flex items-center gap-1 lg:hidden">
            {mounted && !hideThemeToggle ? (
              <button
                onClick={handleToggle}
                className="inline-flex h-10 w-10 items-center justify-center rounded-full transition-all duration-300 hover:bg-muted/60"
                style={{ touchAction: "manipulation" }}
                aria-label={resolvedTheme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
                type="button"
              >
                {resolvedTheme === "dark" ? <SunIcon className="h-4 w-4 text-foreground" /> : <MoonIcon className="h-4 w-4 text-foreground" />}
              </button>
            ) : null}
            <button
              onClick={() => setMobileMenuOpen((value) => !value)}
              className="inline-flex h-10 w-10 items-center justify-center rounded-full transition-all duration-300 hover:bg-muted/60"
              aria-label={mobileMenuOpen ? "Close menu" : "Open menu"}
              type="button"
            >
              {mobileMenuOpen ? <Cross1Icon className="h-4 w-4 text-foreground" /> : <HamburgerMenuIcon className="h-4 w-4 text-foreground" />}
            </button>
          </div>
        </div>
      </nav>

      <div
        className={cn(
          "fixed inset-0 z-40 overflow-y-auto pt-14 transition-all duration-500 ease-out lg:hidden",
          mobileMenuOpen ? "pointer-events-auto opacity-100" : "pointer-events-none opacity-0"
        )}
        onClick={() => setMobileMenuOpen(false)}
      >
        <div
          className={cn(
            "min-h-full bg-background/95 backdrop-blur-2xl transition-transform duration-500 ease-out",
            mobileMenuOpen ? "translate-y-0" : "-translate-y-4"
          )}
        >
          <div className="mx-auto flex max-w-md flex-col px-8 pb-12 pt-8" onClick={(event) => event.stopPropagation()}>
            <div className="space-y-1">
              {navLinks.map((link, index) => (
                <AppLink
                  key={link.href}
                  href={link.href}
                  className={cn(
                    "block py-3 font-serif text-[22px] tracking-[-0.01em] transition-all duration-300",
                    isActive(link.href) ? "font-medium text-foreground" : "text-muted-foreground hover:pl-2 hover:text-foreground"
                  )}
                  style={{ transitionDelay: mobileMenuOpen ? `${index * 40}ms` : "0ms" }}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <span className="flex items-baseline gap-3">
                    {isActive(link.href) ? <span className="relative top-[-2px] inline-block h-[5px] w-[5px] flex-shrink-0 rounded-full bg-primary" /> : null}
                    {link.name}
                  </span>
                </AppLink>
              ))}
            </div>

            {visibleAlbums.length > 0 ? (
              <div className="mt-8 border-t border-border/40 pt-6">
                <span className="text-[11px] font-medium uppercase tracking-[0.12em] text-muted-foreground/60">相册</span>
                <div className="mt-3 grid grid-cols-2 gap-x-4 gap-y-1">
                  {visibleAlbums.map((album) => (
                    <AppLink
                      key={album.id}
                      href={albumHref(album)}
                      className="py-2 text-[13px] text-muted-foreground transition-colors duration-200 hover:text-foreground"
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      {album.name}
                    </AppLink>
                  ))}
                </div>
              </div>
            ) : null}

            <div className="mt-8 border-t border-border/40 pt-6">
              {showPrivateLink ? (
                <AppLink
                  href="/private"
                  className="mb-4 block text-[15px] text-muted-foreground transition-colors duration-200 hover:text-foreground"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  隐私相册
                </AppLink>
              ) : null}
              <AppLink
                href={consoleHref}
                className="text-[15px] text-muted-foreground transition-colors duration-200 hover:text-foreground"
                onClick={() => setMobileMenuOpen(false)}
              >
                {consoleLabel}
              </AppLink>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
