import { CheckCircle2, Clock, AlertTriangle, XCircle, ArrowRightLeft } from 'lucide-react';

const COLORS = {
  PENDING: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  SCHEDULED: 'bg-green-100 text-green-800 border-green-200',
  REDIRECTED: 'bg-blue-100 text-blue-800 border-blue-200',
  DENIED: 'bg-red-100 text-red-800 border-red-200',
  CANCELED: 'bg-gray-100 text-gray-800 border-gray-200',
};

const ICONS = {
  PENDING: Clock,
  SCHEDULED: CheckCircle2,
  REDIRECTED: ArrowRightLeft,
  DENIED: XCircle,
  CANCELED: AlertTriangle,
};

const LABELS = {
  PENDING: 'Pending Review',
  SCHEDULED: 'Scheduled',
  REDIRECTED: 'Redirected',
  DENIED: 'Denied',
  CANCELED: 'Canceled',
};

export const StatusBadge = ({ status, className = '' }) => {
  const Icon = ICONS[status] || Clock;
  const colorClass = COLORS[status] || COLORS.PENDING;
  const label = LABELS[status] || status;

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold tracking-wide border ${colorClass} ${className}`}>
      <Icon className="w-3.5 h-3.5 mr-1" strokeWidth={2.5} />
      {label}
    </span>
  );
};
