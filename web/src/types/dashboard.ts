export interface DashboardStats {
  users: {
    total: number;
    active: number;
    new_today: number;
    new_week: number;
    new_month: number;
  };
  organizations: {
    total: number;
    provinces: number;
    cities: number;
    districts: number;
  };
  missing_persons: {
    total: number;
    missing: number;
    searching: number;
    found: number;
    reunited: number;
    new_today: number;
    new_week: number;
  };
  tasks: {
    total: number;
    pending: number;
    processing: number;
    completed: number;
    overdue: number;
  };
  dialects: {
    total: number;
    featured: number;
    plays: number;
    likes: number;
  };
  files: {
    total_count: number;
    total_size: number;
  };
  recent_activity: Activity[];
}

export interface Activity {
  id: string;
  type: string;
  title: string;
  user_id: string;
  user_name: string;
  entity_type: string;
  entity_id: string;
  created_at: string;
}

export interface TrendData {
  date: string;
  new_cases: number;
  resolved_cases: number;
  new_tasks: number;
  completed_tasks: number;
  new_users: number;
}

export interface OverviewData {
  total_users: number;
  total_cases: number;
  active_cases: number;
  resolved_cases: number;
  pending_tasks: number;
  processing_tasks: number;
}
