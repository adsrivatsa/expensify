interface CategoryBadgeProps {
  icon: string;
  name: string;
  color: string;
}

export function CategoryBadge({ icon, name, color }: CategoryBadgeProps) {
  return (
    <span
      className="badge"
      style={{ backgroundColor: color + '33', borderColor: color }}
      title={name}
    >
      <span>{icon}</span>
      <span>{name}</span>
    </span>
  );
}
