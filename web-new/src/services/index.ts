// API 服务统一导出
export { authApi } from './auth';
export { userApi } from './user';
export { caseApi } from './case';
export { taskApi } from './task';
export { dialectApi } from './dialect';
export { organizationApi } from './organization';
export { dashboardApi } from './dashboard';
export { uploadApi } from './upload';

// 导出类型
export type { LoginRequest, LoginResponse } from './auth';
export type { UpdateUserRequest, ChangePasswordRequest } from './user';
export type { CreateCaseRequest } from './case';
export type { CreateTaskRequest } from './task';
export type { CreateDialectRequest } from './dialect';
export type { CreateOrgRequest } from './organization';
export type { TrendData } from './dashboard';
export type { UploadResponse } from './upload';
