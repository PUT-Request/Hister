export interface HistoryItem {
  id?: number;
  query: string;
  url: string;
  title: string;
  updated_at?: string;
  added?: number;
  updated?: number;
  add_count?: number;
  favicon?: string;
  favicon_key?: string;
  text?: string;
}

export interface DocumentVersion {
  id: number;
  created_at: string;
  html_diff: string;
  text_diff: string;
}
