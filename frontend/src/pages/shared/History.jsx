import { useState, useEffect, useMemo } from 'react';
import { getHistory, getReferralDetails, rescheduleReferral, cancelReferral, downloadAttachment, checkNotifications, markNotificationRead } from '../../services/api';
import { useAuth } from '../../context/AuthContext';
import { StatusBadge } from '../../components/StatusBadge';
import { Calendar, XCircle, Clock, Building2, UserCircle2, CheckCircle2, Copy, Paperclip, File as FileIcon, ExternalLink, Eye, Search, Filter, Bell, Check } from 'lucide-react';

export default function HistoryTab() {
  const { user } = useAuth();
  const [history, setHistory] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedId, setSelectedId] = useState(null);
  const [detail, setDetail] = useState(null);

  // Search and filter state
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  // Rescheduling state
  const [rescheduleData, setRescheduleData] = useState({ id: null, date: '' });

  useEffect(() => {
    loadHistory();
    if (user?.role === 'LEVEL_2_DOC') {
      loadNotifications();
    }
  }, []);

  const loadHistory = async () => {
    setLoading(true);
    try {
      const data = await getHistory();
      setHistory(data.history || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const loadNotifications = async () => {
    try {
      const data = await checkNotifications();
      setNotifications(data.notifications || []);
    } catch (err) {
      console.error('Failed to load notifications:', err);
    }
  };

  const handleMarkNotificationRead = async (id) => {
    try {
      await markNotificationRead(id);
      setNotifications(notifications.map(n =>
        n.id === id ? { ...n, is_read: true } : n
      ));
    } catch (err) {
      console.error('Failed to mark notification read:', err);
    }
  };

  // Filtered history based on search and status
  const filteredHistory = useMemo(() => {
    return history.filter(item => {
      // Search filter - match patient name or CIN
      const searchLower = searchQuery.toLowerCase();
      const matchesSearch = searchQuery === '' ||
        (item.patient_name && item.patient_name.toLowerCase().includes(searchLower)) ||
        (item.patient_cin && item.patient_cin.toLowerCase().includes(searchLower)) ||
        (item.department && item.department.toLowerCase().includes(searchLower));

      // Status filter
      const matchesStatus = statusFilter === 'all' || item.status === statusFilter;

      return matchesSearch && matchesStatus;
    });
  }, [history, searchQuery, statusFilter]);

  const handleCancel = async (id) => {
    if (!window.confirm("Êtes-vous sûr de vouloir annuler ce rendez-vous ?")) return;
    try {
      await cancelReferral(id);
      loadHistory();
    } catch (err) {
      alert("Erreur: " + (err.response?.data?.error || err.message));
    }
  };

  const handleRescheduleSubmit = async (e) => {
    e.preventDefault();
    try {
      const isoDate = new Date(rescheduleData.date).toISOString();
      await rescheduleReferral(rescheduleData.id, isoDate);
      setRescheduleData({ id: null, date: '' });
      loadHistory();
    } catch (err) {
      alert("Erreur: " + (err.response?.data?.error || err.message));
    }
  };

  const calculateAge = (dob) => {
    if (!dob) return 'N/A';
    const birthDate = new Date(dob);
    const today = new Date();
    let age = today.getFullYear() - birthDate.getFullYear();
    const m = today.getMonth() - birthDate.getMonth();
    if (m < 0 || (m === 0 && today.getDate() < birthDate.getDate())) {
      age--;
    }
    return age;
  };

  const handleViewDetail = async (id) => {
    try {
      const data = await getReferralDetails(id);
      setDetail(data);
      setSelectedId(id);
    } catch (err) {
      console.error('Failed to load referral details:', err);
      alert('Impossible de charger les détails du dossier');
    }
  };

  // Helper to get notification icon
  const getNotificationIcon = (message) => {
    if (message.includes('✅')) return <CheckCircle2 className="w-5 h-5 text-green-500 flex-shrink-0" />;
    if (message.includes('❌')) return <XCircle className="w-5 h-5 text-red-500 flex-shrink-0" />;
    if (message.includes('🔄')) return <Calendar className="w-5 h-5 text-blue-500 flex-shrink-0" />;
    return <Bell className="w-5 h-5 text-gray-500 flex-shrink-0" />;
  };

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Notifications Activity Feed - Only for LEVEL_2_DOC */}
      {user?.role === 'LEVEL_2_DOC' && notifications.length > 0 && (
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-100 p-4" role="region" aria-label="Notifications récentes">
          <div className="flex items-center gap-2 mb-3">
            <Bell className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-bold text-gray-900">Activités récentes</h2>
            <span className="bg-blue-100 text-blue-700 text-xs font-bold px-2 py-0.5 rounded-full">
              {notifications.filter(n => !n.is_read).length} non lues
            </span>
          </div>
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {notifications.slice(0, 5).map((notif) => (
              <div
                key={notif.id}
                className={`flex items-start gap-3 p-3 rounded-lg transition-colors ${notif.is_read ? 'bg-white/50 opacity-75' : 'bg-white shadow-sm'
                  }`}
              >
                {getNotificationIcon(notif.message)}
                <div className="flex-1 min-w-0">
                  <p className={`text-sm ${notif.is_read ? 'text-gray-600' : 'text-gray-900 font-medium'}`}>
                    {notif.message.replace(/^[✅❌🔄]\s*/, '')}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">
                    {new Date(notif.created_at).toLocaleString('fr-FR')}
                  </p>
                </div>
                {!notif.is_read && (
                  <button
                    onClick={() => handleMarkNotificationRead(notif.id)}
                    className="text-blue-600 hover:text-blue-800 p-1 rounded hover:bg-blue-100 transition-colors"
                    aria-label="Marquer comme lu"
                    title="Marquer comme lu"
                  >
                    <Check className="w-4 h-4" />
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Search and Filter Controls */}
      <div className="bg-white rounded-xl border border-gray-200 p-4 shadow-sm">
        <div className="flex flex-col md:flex-row gap-4">
          {/* Search Input */}
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Rechercher par nom patient, CIN ou service..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
              aria-label="Rechercher dans l'historique"
            />
          </div>

          {/* Status Filter */}
          <div className="flex items-center gap-2">
            <Filter className="w-5 h-5 text-gray-400" />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors bg-white"
              aria-label="Filtrer par statut"
            >
              <option value="all">Tous les statuts</option>
              <option value="SCHEDULED">Programmé</option>
              <option value="DENIED">Refusé</option>
              <option value="CANCELED">Annulé</option>
              <option value="REDIRECTED">Redirigé</option>
            </select>
          </div>
        </div>
      </div>

      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 border-b-4 border-blue-600 pb-1">
          {user?.role === 'LEVEL_2_DOC' ? 'Historique des envois' : 'Historique des dossiers'}
        </h1>
        <div className="bg-white px-4 py-2 rounded-full shadow-sm text-sm font-medium text-gray-500 border border-gray-200">
          {searchQuery || statusFilter !== 'all' ? (
            <span>
              <span className="font-bold text-gray-900">{filteredHistory.length}</span> / {history.length} résultats
            </span>
          ) : (
            <span>Total: <span className="font-bold text-gray-900">{history.length}</span> enregistrements</span>
          )}
        </div>
      </div>

      {loading ? (
        <div className="text-center p-12 text-gray-500 font-medium animate-pulse">Chargement de l'historique...</div>
      ) : filteredHistory.length === 0 ? (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-12 text-center">
          <Clock className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-bold text-gray-900">
            {searchQuery || statusFilter !== 'all' ? 'Aucun résultat' : 'Aucun historique'}
          </h3>
          <p className="text-gray-500">
            {searchQuery || statusFilter !== 'all'
              ? 'Essayez avec d\'autres critères de recherche.'
              : 'Les dossiers clôs apparaîtront ici.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" role="list" aria-label="Liste des dossiers">
          {filteredHistory.map((item) => (
            <div
              key={item.id}
              className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow flex flex-col"
              role="listitem"
              tabIndex={0}
              aria-label={`Dossier de ${item.patient_name}, statut: ${item.status}`}
            >
              {/* Header */}
              <div className="bg-gray-50 p-4 border-b border-gray-100 flex justify-between items-start">
                <div>
                  <h3 className="font-bold text-gray-900 text-lg flex items-center gap-2">
                    {item.patient_name}
                    <span className="text-xs font-medium text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full" title={item.patient_dob}>
                      {calculateAge(item.patient_dob)} ans
                    </span>
                  </h3>
                  <div className="text-sm font-mono text-gray-400 mt-1" title="CIN Patient (DOB: {item.patient_dob})">{item.patient_cin}</div>
                </div>
                <div className="flex flex-col items-end gap-2">
                  <StatusBadge status={item.status} />
                  {item.has_attachments && (
                    <div className="flex items-center gap-1.5 text-blue-600 bg-blue-50 px-2 py-1 rounded-full text-[10px] font-bold border border-blue-100 animate-pulse">
                      <Paperclip className="w-3 h-3" /> PIÈCES JOINTES
                    </div>
                  )}
                </div>
              </div>

              {/* Body */}
              <div className="p-4 space-y-3 flex-1 flex flex-col">
                <div className="grid grid-cols-2 gap-y-2 text-sm">
                  {user?.role === 'LEVEL_2_DOC' ? (
                    <>
                      <div className="text-gray-500 flex items-center gap-1.5"><Building2 className="w-4 h-4 text-purple-500" /> Service CHU</div>
                      <div className="font-medium text-gray-900 text-right">{item.department}</div>
                    </>
                  ) : (
                    <>
                      <div className="text-gray-500 flex items-center gap-1.5"><Building2 className="w-4 h-4 text-emerald-500" /> Source (Provincial)</div>
                      <div className="font-medium text-gray-900 text-right">{item.creator_facility}</div>
                    </>
                  )}

                  {item.appointment_date && (
                    <>
                      <div className="text-gray-500 flex items-center gap-1.5"><Calendar className="w-4 h-4 text-blue-500" /> RDV Planifié</div>
                      <div className="font-bold text-blue-700 bg-blue-50 px-2 py-0.5 rounded text-right">
                        {new Date(item.appointment_date).toLocaleString('fr-FR', {
                          day: '2-digit', month: '2-digit', year: 'numeric',
                          hour: '2-digit', minute: '2-digit'
                        })}
                      </div>
                    </>
                  )}

                  {item.rejection_reason && (
                    <div className="col-span-2 mt-2 bg-red-50 border border-red-100 p-3 rounded-lg">
                      <p className="text-xs font-bold text-red-800 uppercase tracking-widest mb-1">Motif de refus</p>
                      <p className="text-sm text-red-700 italic">{item.rejection_reason}</p>
                    </div>
                  )}
                </div>

                <div className="mt-4 pt-4 border-t border-gray-100 text-xs text-gray-600 line-clamp-3 italic bg-gray-50/30 p-2 rounded">
                  "{item.symptoms.substring(0, 150)}{item.symptoms.length > 150 ? '...' : ''}"
                </div>

                {/* Attachments List */}
                {item.attachments && item.attachments.length > 0 && (
                  <div className="mt-4 pt-4 border-t border-gray-100">
                    <div className="flex flex-wrap gap-2">
                      {item.attachments.map(att => (
                        <button
                          key={att.id}
                          onClick={async () => {
                            try {
                              const blob = await downloadAttachment(att.id);
                              const url = window.URL.createObjectURL(blob);
                              const a = document.createElement('a');
                              a.href = url;
                              a.download = att.file_name;
                              document.body.appendChild(a);
                              a.click();
                              window.URL.revokeObjectURL(url);
                              document.body.removeChild(a);
                            } catch (err) {
                              console.error('Failed to download attachment:', err);
                            }
                          }}
                          className="flex items-center gap-1.5 bg-gray-50 hover:bg-blue-50 border border-gray-200 hover:border-blue-200 px-2 py-1 rounded text-[10px] font-bold text-gray-700 hover:text-blue-700 transition-colors group cursor-pointer"
                        >
                          <FileIcon className="w-3 h-3 text-gray-400 group-hover:text-blue-500" />
                          <span className="truncate max-w-[80px]">{att.file_name}</span>
                          <ExternalLink className="w-2.5 h-2.5 text-gray-300 group-hover:text-blue-400" />
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <div className="mt-auto pt-4 text-[10px] text-gray-400 font-mono text-center">
                  Dossier créé: {new Date(item.created_at).toLocaleString()}
                </div>

                {/* View Details Button */}
                <button
                  onClick={() => handleViewDetail(item.id)}
                  className="mt-3 w-full flex items-center justify-center gap-1.5 bg-blue-50 text-blue-700 border border-blue-100 py-2 rounded-lg text-sm font-medium hover:bg-blue-100 transition-colors"
                >
                  <Eye className="w-4 h-4" /> Voir le dossier complet
                </button>
              </div>

              {/* CHU_DOC Modification Actions (Only for Scheduled) */}
              {user?.role === 'CHU_DOC' && item.status === 'SCHEDULED' && (
                <div className="bg-gray-50 border-t border-gray-200 p-3">
                  {rescheduleData.id === item.id ? (
                    <form onSubmit={handleRescheduleSubmit} className="flex gap-2">
                      <input
                        type="datetime-local"
                        required
                        value={rescheduleData.date}
                        onChange={(e) => setRescheduleData({ id: item.id, date: e.target.value })}
                        className="flex-1 text-sm border-gray-300 rounded p-1.5"
                      />
                      <button type="submit" className="bg-blue-600 text-white px-3 py-1.5 rounded text-xs font-bold hover:bg-blue-700">OK</button>
                      <button type="button" onClick={() => setRescheduleData({ id: null, date: '' })} className="bg-gray-200 text-gray-700 px-3 py-1.5 rounded text-xs font-bold hover:bg-gray-300">Annuler</button>
                    </form>
                  ) : (
                    <div className="grid grid-cols-2 gap-2">
                      <button
                        onClick={() => setRescheduleData({ id: item.id, date: '' })}
                        className="flex items-center justify-center gap-1.5 w-full bg-white border border-gray-300 text-gray-700 py-2 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors shadow-sm"
                      >
                        <Calendar className="w-4 h-4" /> Modifier RDV
                      </button>
                      <button
                        onClick={() => handleCancel(item.id)}
                        className="flex items-center justify-center gap-1.5 w-full bg-red-50 text-red-700 border border-red-100 py-2 rounded-lg text-sm font-medium hover:bg-red-100 transition-colors"
                      >
                        <XCircle className="w-4 h-4" /> Annuler RDV
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
      {/* Detail Modal */}
      {detail && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
          onClick={(e) => {
            if (e.target === e.currentTarget) {
              setDetail(null);
              setSelectedId(null);
            }
          }}
        >
          <div className="bg-white rounded-2xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            {/* Header */}
            <div className="bg-gradient-to-r from-blue-600 to-blue-700 p-6 rounded-t-2xl">
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="text-2xl font-bold text-white flex items-center gap-3">
                    {detail.patient_name}
                  </h2>
                  <div className="flex flex-wrap items-center mt-2 gap-4 text-sm text-blue-100 font-medium">
                    <span className="flex items-center"><UserCircle2 className="w-4 h-4 mr-1" /> CIN: {detail.patient_cin} / {calculateAge(detail.patient_dob)} ans ({detail.patient_dob})</span>
                    <span className="flex items-center"><Building2 className="w-4 h-4 mr-1" /> Dr. {detail.creator_username} ({detail.creator_facility})</span>
                  </div>
                </div>
                <button
                  onClick={() => { setDetail(null); setSelectedId(null); }}
                  className="text-white/80 hover:text-white hover:bg-white/20 p-2 rounded-lg transition-colors"
                >
                  <XCircle className="w-6 h-6" />
                </button>
              </div>
            </div>

            {/* Status & Info */}
            <div className="p-6 border-b border-gray-100">
              <div className="flex items-center justify-between">
                <div>
                  <StatusBadge status={detail.status} />
                  {detail.appointment_date && (
                    <div className="mt-2 flex items-center gap-2 text-blue-700 font-bold">
                      <Calendar className="w-5 h-5" />
                      RDV: {new Date(detail.appointment_date).toLocaleString('fr-FR', {
                        day: '2-digit', month: '2-digit', year: 'numeric',
                        hour: '2-digit', minute: '2-digit'
                      })}
                    </div>
                  )}
                </div>
                <div className="text-right">
                  <p className="text-xs text-gray-400">Service: {detail.department}</p>
                  <p className="text-xs text-gray-400">Créé: {new Date(detail.created_at).toLocaleString('fr-FR')}</p>
                </div>
              </div>
            </div>

            {/* AI Summary */}
            {detail.ai_summary && (
              <div className="p-6 border-b border-gray-100 bg-blue-50">
                <h3 className="text-sm font-bold text-blue-800 uppercase tracking-wider mb-3">Analyse IA</h3>
                <div className="bg-white rounded-xl border border-blue-100 p-4 shadow-sm">
                  <div className="space-y-2">
                    {detail.ai_summary.split('\n').map((line, i) => (
                      <p key={i} className={`text-sm ${i === 0 ? 'font-bold text-blue-900 text-base' : 'text-gray-700'}`}>
                        {line}
                      </p>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Symptoms */}
            <div className="p-6 border-b border-gray-100">
              <h3 className="text-sm font-bold text-gray-800 uppercase tracking-wider mb-3">Symptoms &amp; Informations</h3>
              <p className="text-gray-700 whitespace-pre-wrap leading-relaxed bg-gray-50 p-4 rounded-xl">
                {detail.symptoms}
              </p>
            </div>

            {/* Rejection Reason */}
            {detail.rejection_reason && (
              <div className="p-6 border-b border-gray-100 bg-red-50">
                <h3 className="text-sm font-bold text-red-800 uppercase tracking-wider mb-3">Motif de refus</h3>
                <p className="text-red-700 italic bg-white p-4 rounded-xl border border-red-100">
                  {detail.rejection_reason}
                </p>
              </div>
            )}

            {/* Attachments */}
            {detail.attachments && detail.attachments.length > 0 && (
              <div className="p-6 border-b border-gray-100">
                <h3 className="text-sm font-bold text-gray-800 uppercase tracking-wider mb-3">
                  <Paperclip className="w-4 h-4 inline mr-1" />
                  Pièces Jointes ({detail.attachments.length})
                </h3>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  {detail.attachments.map((att) => (
                    <button
                      key={att.id}
                      onClick={async () => {
                        try {
                          const blob = await downloadAttachment(att.id);
                          const url = window.URL.createObjectURL(blob);
                          const a = document.createElement('a');
                          a.href = url;
                          a.download = att.file_name;
                          document.body.appendChild(a);
                          a.click();
                          window.URL.revokeObjectURL(url);
                          document.body.removeChild(a);
                        } catch (err) {
                          console.error('Failed to download attachment:', err);
                        }
                      }}
                      className="flex items-center gap-2 bg-gray-50 hover:bg-blue-50 border border-gray-200 hover:border-blue-200 p-3 rounded-lg text-sm transition-colors group"
                    >
                      <FileIcon className="w-5 h-5 text-gray-400 group-hover:text-blue-500" />
                      <span className="truncate text-gray-700 group-hover:text-blue-700">{att.file_name}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Action Buttons for CHU_DOC (Only for Scheduled) */}
            {user?.role === 'CHU_DOC' && detail.status === 'SCHEDULED' && (
              <div className="p-6 bg-gray-50 border-t border-gray-200 rounded-b-2xl">
                {rescheduleData.id === detail.id ? (
                  <form onSubmit={handleRescheduleSubmit} className="flex gap-2 items-center">
                    <input
                      type="datetime-local"
                      required
                      value={rescheduleData.date}
                      onChange={(e) => setRescheduleData({ id: detail.id, date: e.target.value })}
                      className="flex-1 text-sm border-gray-300 rounded p-2"
                    />
                    <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-bold hover:bg-blue-700">Confirmer</button>
                    <button type="button" onClick={() => setRescheduleData({ id: null, date: '' })} className="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg text-sm font-bold hover:bg-gray-300">Annuler</button>
                  </form>
                ) : (
                  <div className="grid grid-cols-2 gap-3">
                    <button
                      onClick={() => setRescheduleData({ id: detail.id, date: '' })}
                      className="flex items-center justify-center gap-2 w-full bg-white border border-gray-300 text-gray-700 py-3 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors shadow-sm"
                    >
                      <Calendar className="w-5 h-5" /> Modifier RDV
                    </button>
                    <button
                      onClick={() => handleCancel(detail.id)}
                      className="flex items-center justify-center gap-2 w-full bg-red-50 text-red-700 border border-red-100 py-3 rounded-lg text-sm font-medium hover:bg-red-100 transition-colors"
                    >
                      <XCircle className="w-5 h-5" /> Annuler RDV
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
