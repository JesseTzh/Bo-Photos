import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { AppProviders } from "./app/providers";
import { router } from "./app/router";
import { SiteIdentity } from "./features/site/site-identity";
import "./styles/index.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AppProviders>
      <SiteIdentity />
      <RouterProvider router={router} />
    </AppProviders>
  </React.StrictMode>
);
