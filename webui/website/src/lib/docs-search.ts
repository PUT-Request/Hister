export interface DocsSearchEntry {
  url: string;
  pageTitle: string;
  description: string;
  heading: string | null;
  content: string;
}

export interface DocsSearchIndex {
  entries: DocsSearchEntry[];
}
