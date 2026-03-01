import { create } from 'zustand';

interface GlobalState {
  // 侧边栏折叠状态
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;

  // 主题
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;

  // 页面加载状态
  pageLoading: boolean;
  setPageLoading: (loading: boolean) => void;

  // 通知数量
  unreadCount: number;
  setUnreadCount: (count: number) => void;
  incrementUnread: () => void;

  // 当前选中组织
  selectedOrgId: string | null;
  setSelectedOrgId: (orgId: string | null) => void;
}

export const useGlobalStore = create<GlobalState>()((set) => ({
  sidebarCollapsed: false,
  toggleSidebar: () =>
    set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

  theme: 'light',
  setTheme: (theme) => set({ theme }),

  pageLoading: false,
  setPageLoading: (loading) => set({ pageLoading: loading }),

  unreadCount: 0,
  setUnreadCount: (count) => set({ unreadCount: count }),
  incrementUnread: () =>
    set((state) => ({ unreadCount: state.unreadCount + 1 })),

  selectedOrgId: null,
  setSelectedOrgId: (orgId) => set({ selectedOrgId: orgId }),
}));
