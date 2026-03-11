import request from './request';
import type {
  PaginatedData,
  MissingPersonResponse,
  MissingPersonListRequest,
  CreateMissingPersonRequest,
  UpdateMissingPersonRequest,
  MissingPersonTrackResponse,
  CreateMissingPersonTrackRequest,
  MarkFoundRequest,
  MissingPersonStatsResponse,
} from '@/types';

export function getMissingPersons(params: MissingPersonListRequest) {
  return request.get<never, PaginatedData<MissingPersonResponse>>('/missing-persons', { params });
}

export function searchMissingPersons(params: MissingPersonListRequest) {
  return request.get<never, PaginatedData<MissingPersonResponse>>('/missing-persons/search', { params });
}

export function getMissingPersonStats() {
  return request.get<never, MissingPersonStatsResponse>('/missing-persons/stats');
}

export function getMissingPerson(id: string) {
  return request.get<never, MissingPersonResponse>(`/missing-persons/${id}`);
}

export function getMissingPersonTracks(id: string) {
  return request.get<never, MissingPersonTrackResponse[]>(`/missing-persons/${id}/tracks`);
}

export function createMissingPerson(data: CreateMissingPersonRequest) {
  return request.post<never, MissingPersonResponse>('/missing-persons', data);
}

export function updateMissingPerson(id: string, data: UpdateMissingPersonRequest) {
  return request.put<never, MissingPersonResponse>(`/missing-persons/${id}`, data);
}

export function deleteMissingPerson(id: string) {
  return request.delete<never, null>(`/missing-persons/${id}`);
}

export function updateMissingPersonStatus(id: string, status: string) {
  return request.put<never, null>(`/missing-persons/${id}/status`, { status });
}

export function markFound(id: string, data: MarkFoundRequest) {
  return request.post<never, null>(`/missing-persons/${id}/found`, data);
}

export function markReunited(id: string) {
  return request.post<never, null>(`/missing-persons/${id}/reunited`);
}

export function addTrack(id: string, data: CreateMissingPersonTrackRequest) {
  return request.post<never, MissingPersonTrackResponse>(`/missing-persons/${id}/tracks`, data);
}
