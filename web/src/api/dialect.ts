import request from './request';
import type {
  PaginatedData,
  DialectResponse,
  DialectListRequest,
  CreateDialectRequest,
  UpdateDialectRequest,
  DialectCommentResponse,
  CreateDialectCommentRequest,
  DialectStatsResponse,
} from '@/types';

export function getDialects(params: DialectListRequest) {
  return request.get<never, PaginatedData<DialectResponse>>('/dialects', { params });
}

export function getFeaturedDialects() {
  return request.get<never, DialectResponse[]>('/dialects/featured');
}

export function getDialectStats() {
  return request.get<never, DialectStatsResponse>('/dialects/stats');
}

export function getDialect(id: string) {
  return request.get<never, DialectResponse>(`/dialects/${id}`);
}

export function getDialectComments(id: string, params?: { page?: number; page_size?: number }) {
  return request.get<never, PaginatedData<DialectCommentResponse>>(`/dialects/${id}/comments`, { params });
}

export function recordPlay(id: string) {
  return request.post<never, null>(`/dialects/${id}/play`);
}

export function likeDialect(id: string) {
  return request.post<never, null>(`/dialects/${id}/like`);
}

export function unlikeDialect(id: string) {
  return request.delete<never, null>(`/dialects/${id}/like`);
}

export function addComment(id: string, data: CreateDialectCommentRequest) {
  return request.post<never, DialectCommentResponse>(`/dialects/${id}/comments`, data);
}

export function createDialect(data: CreateDialectRequest) {
  return request.post<never, DialectResponse>('/dialects', data);
}

export function updateDialect(id: string, data: UpdateDialectRequest) {
  return request.put<never, DialectResponse>(`/dialects/${id}`, data);
}

export function deleteDialect(id: string) {
  return request.delete<never, null>(`/dialects/${id}`);
}

export function updateDialectStatus(id: string, status: string) {
  return request.put<never, null>(`/dialects/${id}/status`, { status });
}

export function featureDialect(id: string) {
  return request.post<never, null>(`/dialects/${id}/feature`);
}

export function unfeatureDialect(id: string) {
  return request.delete<never, null>(`/dialects/${id}/feature`);
}
