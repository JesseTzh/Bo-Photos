import { CameraOutlined } from "@ant-design/icons";
import type { PropsWithChildren } from "react";

export function AuthLayout({ children }: PropsWithChildren) {
  return (
    <main className="auth-shell">
      <section className="auth-intro">
        <div className="brand-mark">
          <CameraOutlined />
        </div>
        <p className="eyebrow">BOPHOTOS</p>
        <h1>让照片回到自己的硬盘，也回到你的视线里。</h1>
        <p className="auth-copy">
          单机、私有、直接。这个入口只属于图库管理员。
        </p>
      </section>
      <section className="auth-panel">{children}</section>
    </main>
  );
}
