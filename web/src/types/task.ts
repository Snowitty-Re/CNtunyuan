import type { PaginationParams } from './api';
import type { UserResponse } from './user';
import type { MissingPersonResponse } from './missingPerson';

export interface TaskResponse {
  id: string;
  title: string;
  description: string;
  type: string;
  priority: string;
  status: string;
  deadline?: string;
  started_at?: string;
  completed_at?: string;
  creator_id: string;
  assignee_id?: string;
  org_id: string;
  missing_person_id?: string;
  location: string;
  province: string;
  city: string;
  district: string;
  address: string;
  lat: number;
  lng: number;
  result: string;
  feedback: string;
  progress: number;
  view_count: number;
  creator?: UserResponse;
  assignee?: UserResponse;
  missing_person?: MissingPersonResponse;
  created_at: string;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  type: string;
  priority?: string;
  deadline?: string;
  missing_person_id?: string;
  location?: string;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  lat?: number;
  lng?: number;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  type?: string;
  priority?: string;
  deadline?: string;
  location?: string;
  province?: string;
  city?: string;
  district?: string;
  address?: string;
  lat?: number;
  lng?: number;
}

export interface TaskListRequest extends PaginationParams {
  keyword?: string;
  type?: string;
  status?: string;
  priority?: string;
  assignee_id?: string;
  missing_person_id?: string;
  is_overdue?: boolean;
}

export interface AssignTaskRequest {
  assignee_id: string;
}

export interface CompleteTaskRequest {
  result?: string;
}

export interface CancelTaskRequest {
  reason?: string;
}

export interface UpdateTaskProgressRequest {
  progress: number;
}

export interface TaskLogResponse {
  id: string;
  task_id: string;
  user_id: string;
  action: string;
  old_status: string;
  new_status: string;
  content: string;
  user?: UserResponse;
  created_at: string;
}

export interface TaskStatsResponse {
  total: number;
  draft: number;
  pending: number;
  assigned: number;
  processing: number;
  completed: number;
  cancelled: number;
  overdue: number;
  my_tasks: number;
  my_pending: number;
  my_completed: number;
}
