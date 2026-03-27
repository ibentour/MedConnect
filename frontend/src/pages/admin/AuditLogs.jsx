import { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { Download, Filter, ChevronLeft, ChevronRight, FileText, FileSpreadsheet } from 'lucide-react';

export default function AuditLogs() {
    const [logs, setLogs] = useState([]);
    const [users, setUsers] = useState([]);
    const [actions, setActions] = useState([]);
    const [loading, setLoading] = useState(true);
    const [pagination, setPagination] = useState({ limit: 50, offset: 0, total: 0, hasNext: false, hasPrev: false });

    // Filter state
    const [filters, setFilters] = useState({
        user_id: '',
        action: '',
        start_date: '',
        end_date: ''
    });

    // Fetch filter options
    useEffect(() => {
        const fetchFilters = async () => {
            try {
                const [usersRes, actionsRes] = await Promise.all([
                    api.get('/admin/audit-logs/users'),
                    api.get('/admin/audit-logs/actions')
                ]);
                setUsers(usersRes.data);
                setActions(actionsRes.data);
            } catch (err) {
                console.error('Failed to load filter options:', err);
            }
        };
        fetchFilters();
    }, []);

    // Fetch logs
    useEffect(() => {
        fetchLogs();
    }, [pagination.offset, pagination.limit]);

    const fetchLogs = async () => {
        setLoading(true);
        try {
            const params = new URLSearchParams();
            if (filters.user_id) params.append('user_id', filters.user_id);
            if (filters.action) params.append('action', filters.action);
            if (filters.start_date) params.append('start_date', filters.start_date);
            if (filters.end_date) params.append('end_date', filters.end_date);
            params.append('limit', pagination.limit);
            params.append('offset', pagination.offset);

            const response = await api.get(`/admin/audit-logs?${params.toString()}`);
            setLogs(response.data.logs);
            setPagination(prev => ({
                ...prev,
                total: response.data.pagination.total,
                hasNext: response.data.pagination.hasNext,
                hasPrev: response.data.pagination.hasPrev
            }));
        } catch (err) {
            console.error('Failed to load audit logs:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleFilterChange = (e) => {
        const { name, value } = e.target;
        setFilters(prev => ({ ...prev, [name]: value }));
    };

    const applyFilters = () => {
        setPagination(prev => ({ ...prev, offset: 0 }));
        fetchLogs();
    };

    const resetFilters = () => {
        setFilters({ user_id: '', action: '', start_date: '', end_date: '' });
        setPagination(prev => ({ ...prev, offset: 0 }));
        fetchLogs();
    };

    const exportData = async (format) => {
        try {
            const params = new URLSearchParams();
            if (filters.start_date) params.append('start_date', filters.start_date);
            if (filters.end_date) params.append('end_date', filters.end_date);
            params.append('format', format);

            const response = await api.get(`/admin/audit-logs/export?${params.toString()}`, {
                responseType: 'blob'
            });

            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', `audit_logs.${format === 'pdf' ? 'html' : 'csv'}`);
            document.body.appendChild(link);
            link.click();
            link.remove();
        } catch (err) {
            console.error('Export failed:', err);
        }
    };

    const goToNextPage = () => {
        if (pagination.hasNext) {
            setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }));
        }
    };

    const goToPrevPage = () => {
        if (pagination.hasPrev) {
            setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }));
        }
    };

    const formatDate = (dateString) => {
        if (!dateString) return '-';
        return new Date(dateString).toLocaleString();
    };

    const getStatusColor = (status) => {
        if (status >= 200 && status < 300) return 'bg-green-100 text-green-800';
        if (status >= 400 && status < 500) return 'bg-yellow-100 text-yellow-800';
        if (status >= 500) return 'bg-red-100 text-red-800';
        return 'bg-gray-100 text-gray-800';
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Audit Logs</h1>
                    <p className="text-sm text-gray-500">View and export system activity logs</p>
                </div>
                <div className="flex gap-2">
                    <button
                        onClick={() => exportData('csv')}
                        className="flex items-center gap-2 px-4 py-2 bg-brand-600 text-white rounded-xl hover:bg-brand-700 transition-colors shadow-sm font-medium"
                    >
                        <FileSpreadsheet className="w-4 h-4" />
                        Export CSV
                    </button>
                    <button
                        onClick={() => exportData('pdf')}
                        className="flex items-center gap-2 px-4 py-2 bg-brand-600 text-white rounded-xl hover:bg-brand-700 transition-colors shadow-sm font-medium"
                    >
                        <FileText className="w-4 h-4" />
                        Export PDF
                    </button>
                </div>
            </div>

            {/* Filters */}
            <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-100">
                <div className="flex items-center gap-2 mb-4">
                    <Filter className="w-5 h-5 text-brand-600" />
                    <h2 className="font-bold text-gray-900">Filtres</h2>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
                        <input
                            type="datetime-local"
                            name="start_date"
                            value={filters.start_date}
                            onChange={handleFilterChange}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-brand-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">End Date</label>
                        <input
                            type="datetime-local"
                            name="end_date"
                            value={filters.end_date}
                            onChange={handleFilterChange}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-brand-500"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">User</label>
                        <select
                            name="user_id"
                            value={filters.user_id}
                            onChange={handleFilterChange}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-brand-500"
                        >
                            <option value="">All Users</option>
                            {users.map(user => (
                                <option key={user.id} value={user.id}>{user.username}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Action</label>
                        <select
                            name="action"
                            value={filters.action}
                            onChange={handleFilterChange}
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-brand-500 focus:border-brand-500"
                        >
                            <option value="">All Actions</option>
                            {actions.map(action => (
                                <option key={action} value={action}>{action}</option>
                            ))}
                        </select>
                    </div>
                    <div className="flex items-end gap-2">
                        <button
                            onClick={applyFilters}
                            className="flex-1 px-4 py-2.5 bg-brand-600 text-white rounded-xl hover:bg-brand-700 transition-colors shadow-sm font-bold text-sm"
                        >
                            Appliquer
                        </button>
                        <button
                            onClick={resetFilters}
                            className="px-4 py-2.5 border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors text-sm font-medium text-gray-600"
                        >
                            Réinitialiser
                        </button>
                    </div>
                </div>
            </div>

            {/* Logs Table */}
            <div className="bg-white rounded-xl shadow-sm overflow-hidden border border-gray-100">
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-gray-50">
                            <tr>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Target ID</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">IP Address</th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                            {loading ? (
                                <tr>
                                    <td colSpan={6} className="px-4 py-8 text-center text-gray-500">Loading...</td>
                                </tr>
                            ) : logs.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-4 py-8 text-center text-gray-500">No audit logs found</td>
                                </tr>
                            ) : (
                                logs.map(log => (
                                    <tr key={log.id} className="hover:bg-gray-50">
                                        <td className="px-4 py-3 text-sm text-gray-900 whitespace-nowrap">
                                            {formatDate(log.timestamp)}
                                        </td>
                                        <td className="px-4 py-3 text-sm text-gray-900">
                                            {log.username || '-'}
                                        </td>
                                        <td className="px-4 py-3 text-sm text-gray-600 font-mono">
                                            {log.action}
                                        </td>
                                        <td className="px-4 py-3 text-sm text-gray-500 font-mono">
                                            {log.target_id || '-'}
                                        </td>
                                        <td className="px-4 py-3 text-sm text-gray-500 font-mono">
                                            {log.ip_address}
                                        </td>
                                        <td className="px-4 py-3">
                                            <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(log.status)}`}>
                                                {log.status}
                                            </span>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                <div className="px-6 py-4 bg-gray-50/50 border-t border-gray-100 flex items-center justify-between">
                    <div className="text-sm text-gray-700">
                        Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, pagination.total)} of {pagination.total} results
                    </div>
                    <div className="flex gap-2">
                        <button
                            onClick={goToPrevPage}
                            disabled={!pagination.hasPrev}
                            className="flex items-center gap-1 px-3 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium transition-colors"
                        >
                            <ChevronLeft className="w-4 h-4" />
                            Précédent
                        </button>
                        <button
                            onClick={goToNextPage}
                            disabled={!pagination.hasNext}
                            className="flex items-center gap-1 px-3 py-2 border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium transition-colors"
                        >
                            Next
                            <ChevronRight className="w-4 h-4" />
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
