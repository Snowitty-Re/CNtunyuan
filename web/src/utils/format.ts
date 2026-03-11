import dayjs from 'dayjs';

export function formatDate(date: string | undefined, fmt = 'YYYY-MM-DD'): string {
  if (!date) return '-';
  return dayjs(date).format(fmt);
}

export function formatDateTime(date: string | undefined): string {
  return formatDate(date, 'YYYY-MM-DD HH:mm:ss');
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function maskPhone(phone: string): string {
  if (!phone || phone.length < 7) return phone;
  return phone.replace(/(\d{3})\d{4}(\d+)/, '$1****$2');
}

export function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, '0')}`;
}
