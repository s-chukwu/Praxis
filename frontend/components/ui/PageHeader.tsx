export default function PageHeader({ title, description }: { title: string; description: string }) {
  return (
    <div>
      <h1 className="text-xl font-semibold text-white">{title}</h1>
      <p className="mt-1 text-sm text-gray-400">{description}</p>
    </div>
  );
}