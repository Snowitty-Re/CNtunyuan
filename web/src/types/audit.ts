import type { AuditLogType, AuditLogStatus } from './enums';

export interface AuditLogResponse {
  id: string;
  user_id: string;
  username: string;
  user_role: string;
  module: string;
  action: string;
  type: AuditLogType;
  description: string;
  request_method: string;
  request_url: string;
  request_ip: string;
  request_body: string;
  response_code: number;
  response_body: string;
  user_agent: string;
  duration_ms: number;
  status: AuditLogStatus;
  error_message: string;
  trace_id: string;
  created_at: string;
}

export interface AuditLogListRequest {
  page?: number;
  page_size?: number;
  user_id?: string;
  module?: string;
  action?: string;
  type?: string;
  status?: string;
  start_time?: string;
  end_time?: string;
  keyword?: string;
  request_ip?: string;
}

export interface AuditLogStats {
  total_count: number;
  today_count: number;
  login_count: number;
  operation_count: number;
  error_count: number;
}

export interface AuditUserActivity {
  date: string;
  count: number;
}

export interface AuditModuleStat {
  module: string;
  count: number;
}
