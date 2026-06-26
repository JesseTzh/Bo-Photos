import { Button, Result, Skeleton } from "antd";
import { useNavigate, useParams } from "react-router-dom";
import { useAlbums } from "../features/albums/api";
import { AlbumSortView } from "../features/albums/album-sort-view";

export function AdminAlbumSortPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const albums = useAlbums(true);
  const album = albums.data?.items.find((item) => item.id === id);
  if (albums.isPending) return <Skeleton active />;
  if (!album) return <Result status="404" title="相册不存在" extra={<Button onClick={() => navigate("/admin/albums")}>返回相册</Button>} />;
  return <AlbumSortView album={album} onBack={() => navigate("/admin/albums")} />;
}
