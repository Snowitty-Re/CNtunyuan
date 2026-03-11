export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
  trace_id?: string;
}

export interface PaginatedData<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface PaginationParams {
  page?: number;
  page_size?: number;
}
