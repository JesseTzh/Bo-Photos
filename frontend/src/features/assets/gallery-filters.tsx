import { Select } from "antd";

interface GalleryFiltersProps {
  cameras: string[];
  lenses: string[];
  tags: Array<{ value: string; label: string }>;
  selectedCameras: string[];
  selectedLenses: string[];
  selectedTags: string[];
  tagsOperator: "and" | "or";
  sort?: "asc" | "desc";
  onCamerasChange: (values: string[]) => void;
  onLensesChange: (values: string[]) => void;
  onTagsChange: (values: string[]) => void;
  onTagsOperatorChange: (value: "and" | "or") => void;
  onSortChange: (value?: "asc" | "desc") => void;
}

export function GalleryFilters({
  cameras,
  lenses,
  tags,
  selectedCameras,
  selectedLenses,
  selectedTags,
  tagsOperator,
  sort,
  onCamerasChange,
  onLensesChange,
  onTagsChange,
  onTagsOperatorChange,
  onSortChange
}: GalleryFiltersProps) {
  return (
    <div className="gallery-filters">
      <Select
        mode="multiple"
        allowClear
        placeholder="相机"
        value={selectedCameras}
        options={cameras.map((value) => ({ value, label: value }))}
        onChange={onCamerasChange}
      />
      <Select
        mode="multiple"
        allowClear
        placeholder="镜头"
        value={selectedLenses}
        options={lenses.map((value) => ({ value, label: value }))}
        onChange={onLensesChange}
      />
      <Select
        mode="multiple"
        allowClear
        placeholder="标签"
        value={selectedTags}
        options={tags}
        onChange={onTagsChange}
      />
      <Select
        value={tagsOperator}
        options={[
          { value: "and", label: "包含全部标签" },
          { value: "or", label: "包含任一标签" }
        ]}
        onChange={onTagsOperatorChange}
      />
      <Select
        allowClear
        placeholder="拍摄时间"
        value={sort}
        options={[
          { value: "desc", label: "拍摄时间：新到旧" },
          { value: "asc", label: "拍摄时间：旧到新" }
        ]}
        onChange={onSortChange}
      />
    </div>
  );
}
