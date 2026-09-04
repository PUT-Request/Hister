const modules = import.meta.glob('../../content/datasets/*.json', { eager: true });

export interface Dataset {
  slug: string;
  name: string;
  description: string;
  downloadUrl: string;
  image: string | null;
  tags: string[];
  license: string;
  author: string;
  diskSizeBytes: number | null;
  documentCount: number | null;
  latestUpdate: string | null;
}

export async function load() {
  const datasets: Dataset[] = Object.entries(modules).map(([path, mod]) => {
    const slug = path.split('/').pop()?.replace('.json', '') ?? path;
    const data = (mod as { default: unknown }).default as {
      name: string;
      description: string;
      downloadUrl: string;
      image?: string | null;
      tags?: string[];
      license: string;
      author: string;
      diskSizeBytes?: number | null;
      documentCount?: number | null;
      latestUpdate?: string | null;
    };
    return {
      slug,
      name: data.name,
      description: data.description,
      downloadUrl: data.downloadUrl,
      image: data.image ?? null,
      tags: data.tags ?? [],
      license: data.license,
      author: data.author,
      diskSizeBytes: data.diskSizeBytes ?? null,
      documentCount: data.documentCount ?? null,
      latestUpdate: data.latestUpdate ?? null,
    };
  });

  datasets.sort((a, b) => a.name.localeCompare(b.name));

  return { datasets };
}
