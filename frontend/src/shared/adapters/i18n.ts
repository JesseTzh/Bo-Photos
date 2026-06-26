const dictionary: Record<string, string> = {
  "Link.dashboard": "控制台",
  "Login.signIn": "登录",
  "Button.loading": "加载中...",
  "Button.prev": "上一张",
  "Button.next": "下一张",
  "Button.goBack": "返回",
  "Tips.noImg": "暂无图片",
  "Tips.loadFail": "加载失败，点击重试",
  "Preview.untitled": "未命名",
  "Preview.copyLink": "复制直链",
  "Preview.shareLink": "分享链接",
  "Preview.download": "下载",
  "Preview.fullscreen": "全屏",
  "Exif.title": "EXIF",
  "Exif.camera": "相机",
  "Exif.lens": "镜头",
  "Exif.date": "日期",
  "Exif.aperture": "光圈",
  "Exif.shutter": "快门",
  "Exif.focalLength": "焦距",
  "Exif.iso": "ISO",
  "Exif.resolution": "尺寸",
  "Exif.location": "位置",
  "Words.explore_now": "立即探索"
};

export function useTranslations(namespace?: string) {
  return (key: string, values?: Record<string, string | number>) => {
    const fullKey = namespace ? `${namespace}.${key}` : key;
    let text = dictionary[fullKey] ?? dictionary[key] ?? key;
    if (values) {
      for (const [name, value] of Object.entries(values)) {
        text = text.replace(`{${name}}`, String(value));
      }
    }
    return text;
  };
}
