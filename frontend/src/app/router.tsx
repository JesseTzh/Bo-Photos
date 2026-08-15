import { createBrowserRouter, Navigate } from "react-router-dom";
import { AdminRoute, LoginRoute, PrivateRoute } from "../features/auth/guards";
import { AdminPage } from "../pages/admin-page";
import { HomePage } from "../pages/home-page";
import { LoginPage } from "../pages/login-page";
import { GalleryPage } from "../pages/gallery-page";
import { PreviewPage } from "../pages/preview-page";
import { AdminAssetsPage } from "../pages/admin-assets-page";
import { AdminDashboardPage } from "../pages/admin-dashboard-page";
import { AdminAlbumsPage } from "../pages/admin-albums-page";
import { AdminAlbumSortPage } from "../pages/admin-album-sort-page";
import { AdminAlbumEditPage } from "../pages/admin-album-edit-page";
import { AdminAlbumImagesPage } from "../pages/admin-album-images-page";
import { AdminTagsPage } from "../pages/admin-tags-page";
import { AlbumsPage } from "../pages/albums-page";
import { GuidesPage } from "../pages/guides-page";
import { GuideDetailPage } from "../pages/guide-detail-page";
import { AdminGuidesPage } from "../pages/admin-guides-page";
import { AdminGuideEditorPage } from "../pages/admin-guide-editor-page";
import { AboutPage } from "../pages/about-page";
import { AdminSettingsPage } from "../pages/admin-settings-page";
import { AdminAnalyticsPage } from "../pages/admin-analytics-page";
import { AdminAccountPage } from "../pages/admin-account-page";
import { PrivateAlbumPage } from "../pages/private-album-page";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <HomePage />
  },
  {
    path: "/gallery",
    element: <GalleryPage />
  },
  {
    path: "/gallery/:album",
    element: <GalleryPage />
  },
  {
    path: "/albums",
    element: <AlbumsPage />
  },
  {
    path: "/covers",
    element: <AlbumsPage coversOnly />
  },
  {
    path: "/guides",
    element: <GuidesPage />
  },
  {
    path: "/guides/:id",
    element: <GuideDetailPage />
  },
  { path: "/about", element: <AboutPage /> },
  {
    path: "/preview/:id",
    element: <PreviewPage />
  },
  {
    path: "/private",
    element: (
      <PrivateRoute>
        <PrivateAlbumPage />
      </PrivateRoute>
    )
  },
  {
    path: "/setup",
    element: <Navigate to="/login" replace />
  },
  {
    path: "/:album",
    element: <GalleryPage />
  },
  {
    path: "/login",
    element: (
      <LoginRoute>
        <LoginPage />
      </LoginRoute>
    )
  },
  {
    path: "/admin/*",
    element: (
      <AdminRoute>
        <AdminPage />
      </AdminRoute>
    ),
    children: [
      { index: true, element: <AdminDashboardPage /> },
      { path: "images", element: <AdminAssetsPage /> },
      { path: "albums", element: <AdminAlbumsPage /> },
      { path: "albums/new", element: <AdminAlbumEditPage /> },
      { path: "albums/:id/edit", element: <AdminAlbumEditPage /> },
      { path: "albums/:id/images", element: <AdminAlbumImagesPage /> },
      { path: "albums/:id/sort", element: <AdminAlbumSortPage /> },
      { path: "tags", element: <AdminTagsPage /> },
      { path: "guides", element: <AdminGuidesPage /> },
      { path: "guides/:id", element: <AdminGuideEditorPage /> },
      { path: "analytics", element: <AdminAnalyticsPage /> },
      { path: "settings", element: <AdminSettingsPage /> },
      { path: "account", element: <AdminAccountPage /> }
    ]
  }
]);
