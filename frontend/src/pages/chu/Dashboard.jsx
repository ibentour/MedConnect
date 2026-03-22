import { useState, useEffect } from 'react';
import { getQueue, getReferralDetails, scheduleReferral, redirectReferral, denyReferral, getDirectory, getAttachmentUrl } from '../../services/api';
import { subscribe as wsSubscribe, WS_EVENTS } from '../../services/websocket';
import { AntiLeak } from '../../components/Security';
import { StatusBadge } from '../../components/StatusBadge';
import { Calendar, XCircle, X, ArrowRightLeft, Clock, Activity, AlertTriangle, Building2, UserCircle2, CheckCircle2, File as FileIcon, Paperclip, ExternalLink, Bell } from 'lucide-react';

export default function ChuDashboard() {
  const [queue, setQueue] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [selectedId, setSelectedId] = useState(null);
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);

  // Modals / Dropdowns
  const [showSchedule, setShowSchedule] = useState(false);
  const [showRedirect, setShowRedirect] = useState(false);
  const [showDeny, setShowDeny] = useState(false);
  const [actionData, setActionData] = useState({ date: '', time: '', reason: '', newDeptId: '' });
  const [notification, setNotification] = useState(null);
  const [previewImage, setPreviewImage] = useState(null);

  useEffect(() => {
    loadQueue();
    loadDirectory();

    // Subscribe to WebSocket notifications
    const unsubscribe = wsSubscribe((message) => {
      if (message.type === WS_EVENTS.NEW_REFERRAL) {
        // New urgent referral received - refresh queue
        loadQueue();
        // Show notification
        setNotification({
          type: 'new_referral',
          message: `Nouvelle demande urgente: ${message.patient_name || 'Patient'} (${message.urgency})`,
          referralId: message.referral_id
        });
        // Auto-dismiss after 5 seconds
        setTimeout(() => setNotification(null), 5000);
      } else if (message.type === WS_EVENTS.REFERRAL_REDIRECT) {
        // Referral was redirected - refresh queue
        loadQueue();
        if (selectedId === message.referral_id) {
          setSelectedId(null);
          setDetail(null);
        }
      }
    });

    return () => unsubscribe();
  }, []);

  useEffect(() => {
    if (selectedId) loadDetail(selectedId);
  }, [selectedId]);

  const loadQueue = async () => {
    try {
      const data = await getQueue();
      setQueue(data.queue || []);
    } catch (err) {
      setError('Impossible de charger la file d\'attente. ' + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const loadDirectory = async () => {
    try {
      const data = await getDirectory();
      setDepartments(data.departments);
    } catch (err) {
      console.error(err);
    }
  };

  const loadDetail = async (id) => {
    setDetail(null);
    try {
      const data = await getReferralDetails(id);
      setDetail(data);
    } catch (err) {
      setError('Impossible de charger les détails du dossier.');
    }
  };

  // ── Actions ──────────────────────────────────────────────

  const handleSchedule = async (e) => {
    e.preventDefault();
    setActionLoading(true);
    try {
      // Construct ISO string
      const isoDate = new Date(`${actionData.date}T${actionData.time}:00`).toISOString();
      await scheduleReferral(selectedId, isoDate);
      setShowSchedule(false);
      setDetail({ ...detail, status: 'SCHEDULED', appointment_date: isoDate });

      // Update queue item
      setQueue(queue.map(q => q.id === selectedId ? { ...q, status: 'SCHEDULED' } : q));
    } catch (err) {
      alert(err.response?.data?.error || "Erreur de planification");
    } finally {
      setActionLoading(false);
    }
  };

  const handleRedirect = async (e) => {
    e.preventDefault();
    setActionLoading(true);
    try {
      await redirectReferral(selectedId, actionData.newDeptId, actionData.reason);
      setShowRedirect(false);
      // Remove from queue since it belongs to another dept now
      setQueue(queue.filter(q => q.id !== selectedId));
      setDetail(null);
      setSelectedId(null);
    } catch (err) {
      alert(err.response?.data?.error || "Erreur de réorientation");
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeny = async (e) => {
    e.preventDefault();
    setActionLoading(true);
    try {
      await denyReferral(selectedId, actionData.reason);
      setShowDeny(false);
      setDetail({ ...detail, status: 'DENIED', rejection_reason: actionData.reason });
      // Remove from active pending queue
      setQueue(queue.filter(q => q.id !== selectedId));
    } catch (err) {
      alert(err.response?.data?.error || "Erreur de refus");
    } finally {
      setActionLoading(false);
    }
  };

  // ── UI Helpers ──────────────────────────────────────────

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

  const getUrgencyColor = (urgency) => {
    switch (urgency) {
      case 'CRITICAL': return 'bg-red-500';
      case 'HIGH': return 'bg-orange-500';
      case 'MEDIUM': return 'bg-yellow-500';
      default: return 'bg-green-500';
    }
  };

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col md:flex-row gap-6">
      {/* Notification Toast */}
      {notification && (
        <div className="fixed top-20 right-4 z-50 animate-in slide-in-from-right-2 fade-in duration-300">
          <div className="bg-gradient-to-r from-chu-500 to-chu-600 text-white px-5 py-4 rounded-xl shadow-lg border border-chu-400 max-w-sm">
            <div className="flex items-start gap-3">
              <div className="bg-white/20 p-2 rounded-lg">
                <Bell className="w-5 h-5 text-white" />
              </div>
              <div className="flex-1">
                <h4 className="font-bold text-sm">Nouvelle demande urgente!</h4>
                <p className="text-xs text-white/90 mt-1">{notification.message}</p>
              </div>
              <button
                onClick={() => setNotification(null)}
                className="text-white/70 hover:text-white"
              >
                <XCircle className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Left Pane: Master Queue ── */}
      <div className="w-full md:w-1/3 bg-white rounded-2xl shadow-sm border border-gray-200 flex flex-col overflow-hidden">
        <div className="p-4 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
          <h2 className="text-lg font-bold text-gray-900 flex items-center">
            <Activity className="w-5 h-5 mr-2 text-chu-500" />
            File d'Attente
          </h2>
          <span className="bg-chu-100 text-chu-800 text-xs font-bold px-2 py-1 rounded-full">
            {queue.length} Total
          </span>
        </div>

        {loading ? (
          <div className="flex-1 flex justify-center items-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-chu-500"></div>
          </div>
        ) : queue.length === 0 ? (
          <div className="flex-1 flex flex-col justify-center items-center p-6 text-gray-400">
            <CheckCircle2 className="w-12 h-12 mb-2 opacity-50" />
            <p>La file d'attente est vide.</p>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-3 space-y-3">
            {queue.map((item) => (
              <div
                key={item.id}
                onClick={() => setSelectedId(item.id)}
                className={`p-4 rounded-xl cursor-pointer border hover:border-chu-300 transition-all ${selectedId === item.id
                  ? 'bg-chu-50 border-chu-400 shadow-sm ring-1 ring-chu-400'
                  : 'bg-white border-gray-200'
                  }`}
              >
                <div className="flex justify-between items-start mb-2">
                  <div className="flexItems-center space-x-2">
                    <div className={`w-2.5 h-2.5 rounded-full ${getUrgencyColor(item.urgency)} mt-1 flex-shrink-0`}></div>
                    <AntiLeak><h3 className="font-bold text-gray-900 line-clamp-1">{item.patient_name || 'Patient Anonyme'}</h3></AntiLeak>
                  </div>
                  <span className="text-xs font-medium text-gray-500 whitespace-nowrap bg-gray-100 px-2 py-0.5 rounded">
                    {calculateAge(item.patient_dob)} ans
                  </span>
                </div>

                <p className="text-xs text-brand-600 font-bold mb-2 truncate">
                  <Building2 className="inline w-3 h-3 mr-1" />
                  {item.creator_facility}
                </p>

                {item.ai_summary && (
                  <div className="text-xs text-gray-600 mb-3 bg-gray-50 p-2 rounded line-clamp-2 italic border-l-2 border-chu-200">
                    "{item.ai_summary.split('\n')[0]}"
                  </div>
                )}

                <div className="flex justify-between items-center text-[10px] text-gray-400">
                  <span className="font-semibold text-gray-500">{item.urgency}</span>
                  <div className="flex items-center gap-2">
                    {item.has_attachments && <Paperclip className="w-3.5 h-3.5 text-chu-500" />}
                    <StatusBadge status={item.status} className="scale-90 origin-right" />
                  </div>
                  <span className="hidden xl:inline">{new Date(item.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ── Right Pane: Detail View ── */}
      <div className="w-full md:w-2/3 bg-white rounded-2xl shadow-sm border border-gray-200 flex flex-col overflow-hidden">
        {error && (
          <div className="m-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm font-medium border border-red-200 flex items-center">
            <AlertTriangle className="w-4 h-4 mr-2" /> {error}
          </div>
        )}

        {!selectedId ? (
          <div className="flex-1 flex flex-col justify-center items-center text-gray-400 p-8 text-center bg-gray-50/50">
            <Activity className="w-16 h-16 mb-4 text-gray-200" />
            <h3 className="text-lg font-medium text-gray-900 mb-1">Aucun dossier sélectionné</h3>
            <p>Sélectionnez une demande dans la liste de gauche pour l'analyser et la traiter.</p>
          </div>
        ) : !detail ? (
          <div className="flex-1 flex justify-center items-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-chu-500"></div>
          </div>
        ) : (
          <>
            {/* Header */}
            <div className="p-6 border-b border-gray-200 flex justify-between items-start bg-white">
              <div>
                <AntiLeak>
                  <h2 className="text-2xl font-bold text-gray-900 flex items-center gap-3">
                    {detail.patient_name}
                    <span className={`inline-block w-3 h-3 rounded-full ${getUrgencyColor(detail.urgency)}`} title={`Urgence: ${detail.urgency}`}></span>
                  </h2>
                  <div className="flex flex-wrap items-center mt-2 gap-4 text-sm text-gray-600 font-medium">
                    <span className="flex items-center"><UserCircle2 className="w-4 h-4 mr-1 text-gray-400" /> CIN: {detail.patient_cin} / {calculateAge(detail.patient_dob)} ans ({detail.patient_dob})</span>
                    <span className="flex items-center"><Building2 className="w-4 h-4 mr-1 text-gray-400" /> Dr. {detail.creator_username} ({detail.creator_facility})</span>
                  </div>
                </AntiLeak>
              </div>
              <StatusBadge status={detail.status} />
            </div>

            {/* Scrollable Content */}
            <div className="flex-1 p-6 overflow-y-auto space-y-8 bg-gray-50/30">

              {/* AI Summary Banner */}
              {detail.ai_summary && (
                <div className="bg-blue-50 border border-blue-200 rounded-xl p-5 shadow-sm">
                  <h3 className="text-xs font-bold text-blue-800 uppercase tracking-wider flex items-center mb-3">
                    <Activity className="w-4 h-4 mr-1" /> Synthèse Clinique IA
                  </h3>
                  <AntiLeak>
                    <div className="space-y-2">
                      {detail.ai_summary.split('\n').map((line, i) => (
                        <p key={i} className={`text-sm ${i === 0 ? 'font-bold text-blue-900 text-base' :
                          i === 1 ? 'text-blue-800' : 'text-blue-700 italic border-t border-blue-200/50 pt-2 mt-2'
                          }`}>
                          {line}
                        </p>
                      ))}
                    </div>
                  </AntiLeak>
                </div>
              )}

              {/* Full Symptoms Box */}
              <div className="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm">
                <div className="px-5 py-3 border-b border-gray-100 bg-gray-50">
                  <h3 className="text-sm font-bold text-gray-800">Dossier Clinique Complet</h3>
                </div>
                <div className="p-5">
                  <AntiLeak>
                    <p className="text-gray-700 whitespace-pre-wrap leading-relaxed">
                      {detail.symptoms}
                    </p>
                  </AntiLeak>
                </div>
              </div>

              {/* Attachments Section */}
              {detail.attachments && detail.attachments.length > 0 && (
                <div className="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm">
                  <div className="px-5 py-3 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                    <h3 className="text-sm font-bold text-gray-800 flex items-center gap-2">
                      <Paperclip className="w-4 h-4 text-chu-500" />
                      Pièces Jointes & Documents ({detail.attachments.length})
                    </h3>
                  </div>
                  <div className="p-5">
                    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                      {detail.attachments.map((att) => (
                        <button
                          key={att.id}
                          onClick={() => {
                            const token = localStorage.getItem('token');
                            const url = `${import.meta.env.VITE_API_URL || 'http://localhost:3000/api'}/attachments/${att.id}?token=${encodeURIComponent(token)}`;
                            // Check if it's an image - show inline preview
                            if (att.file_type && att.file_type.startsWith('image/')) {
                              setPreviewImage({ url, fileName: att.file_name });
                            } else {
                              // For non-images, open in new tab
                              window.open(url, '_blank');
                            }
                          }}
                          className="group relative bg-gray-50 rounded-xl border border-gray-100 p-4 transition-all hover:bg-chu-50 hover:border-chu-200 flex flex-col items-center justify-center text-center gap-2 cursor-pointer"
                        >
                          <div className="w-12 h-12 bg-white rounded-lg shadow-inner flex items-center justify-center text-chu-500 group-hover:scale-110 transition-transform">
                            <FileIcon className="w-6 h-6" />
                          </div>
                          <div className="flex flex-col min-w-0 w-full">
                            <span className="text-[10px] font-bold text-gray-700 truncate px-1" title={att.file_name}>
                              {att.file_name}
                            </span>
                            <span className="text-[8px] text-gray-400 font-bold uppercase">
                              {(att.file_size / 1024).toFixed(1)} KB • {att.file_type.split('/')[1]}
                            </span>
                          </div>
                          <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                            <ExternalLink className="w-3 h-3 text-chu-400" />
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                </div>
              )}

            </div>

            {/* Sticky Action Footer */}
            {(detail.status === 'PENDING' || detail.status === 'REDIRECTED') && (
              <div className="p-4 border-t border-gray-200 bg-white shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)]">

                {/* Regular Action Buttons */}
                {(!showSchedule && !showRedirect && !showDeny) && (
                  <div className="flex flex-wrap gap-3 justify-end">
                    <button onClick={() => setShowDeny(true)} className="px-4 py-2 border-2 border-red-200 text-red-600 font-bold rounded-lg hover:bg-red-50 hover:border-red-300 transition-colors flex items-center">
                      <XCircle className="w-4 h-4 mr-2" /> Refuser
                    </button>
                    <button onClick={() => setShowRedirect(true)} className="px-4 py-2 border-2 border-blue-200 text-blue-600 font-bold rounded-lg hover:bg-blue-50 hover:border-blue-300 transition-colors flex items-center">
                      <ArrowRightLeft className="w-4 h-4 mr-2" /> Réorienter
                    </button>
                    <button onClick={() => setShowSchedule(true)} className="px-6 py-2 bg-green-600 text-white font-bold rounded-lg hover:bg-green-700 shadow-md flex items-center transition-colors">
                      <Calendar className="w-4 h-4 mr-2" /> Planifier RDV
                    </button>
                  </div>
                )}

                {/* Schedule Inline Form */}
                {showSchedule && (
                  <form onSubmit={handleSchedule} className="bg-green-50 p-4 rounded-xl border border-green-200 animate-in slide-in-from-bottom-2">
                    <h4 className="font-bold text-green-900 mb-3 flex items-center">
                      <Calendar className="w-4 h-4 mr-2" /> Fixer le Rendez-vous (Génère WhatsApp)
                    </h4>
                    <div className="flex gap-4 items-end">
                      <div className="flex-1">
                        <label className="block text-xs font-bold text-green-800 mb-1">Date</label>
                        <input type="date" required min={new Date().toISOString().split('T')[0]} value={actionData.date} onChange={e => setActionData({ ...actionData, date: e.target.value })} className="w-full rounded border-green-300 focus:border-green-500 focus:ring-green-500 p-2 text-sm" />
                      </div>
                      <div className="flex-1">
                        <label className="block text-xs font-bold text-green-800 mb-1">Heure</label>
                        <input type="time" required value={actionData.time} onChange={e => setActionData({ ...actionData, time: e.target.value })} className="w-full rounded border-green-300 focus:border-green-500 focus:ring-green-500 p-2 text-sm" />
                      </div>
                      <div className="flex gap-2">
                        <button type="button" onClick={() => setShowSchedule(false)} className="px-4 py-2 text-green-700 bg-white border border-green-300 rounded font-medium hover:bg-green-100 text-sm">Annuler</button>
                        <button type="submit" disabled={actionLoading} className="px-6 py-2 bg-green-600 text-white rounded font-bold hover:bg-green-700 text-sm shadow-sm disabled:opacity-50">Confirmer</button>
                      </div>
                    </div>
                  </form>
                )}

                {/* Redirect Inline Form */}
                {showRedirect && (
                  <form onSubmit={handleRedirect} className="bg-blue-50 p-4 rounded-xl border border-blue-200 animate-in slide-in-from-bottom-2">
                    <h4 className="font-bold text-blue-900 mb-3 flex items-center">
                      <ArrowRightLeft className="w-4 h-4 mr-2" /> Réorienter le dossier
                    </h4>
                    <div className="space-y-3">
                      <select required value={actionData.newDeptId} onChange={e => setActionData({ ...actionData, newDeptId: e.target.value })} className="w-full rounded border-blue-300 p-2 text-sm text-gray-900">
                        <option value="">Sélectionner le nouveau service...</option>
                        {departments.filter(d => d.id !== detail.department_id && d.is_accepting).map(d => (
                          <option key={d.id} value={d.id}>{d.name}</option>
                        ))}
                      </select>
                      <input type="text" placeholder="Raison de la réorientation (obligatoire)" required minLength={5} value={actionData.reason} onChange={e => setActionData({ ...actionData, reason: e.target.value })} className="w-full rounded border-blue-300 p-2 text-sm" />
                      <div className="flex justify-end gap-2 pt-2">
                        <button type="button" onClick={() => setShowRedirect(false)} className="px-4 py-2 text-blue-700 bg-white border border-blue-300 rounded font-medium hover:bg-blue-100 text-sm">Annuler</button>
                        <button type="submit" disabled={actionLoading} className="px-6 py-2 bg-blue-600 text-white rounded font-bold hover:bg-blue-700 text-sm shadow-sm disabled:opacity-50">Transférer</button>
                      </div>
                    </div>
                  </form>
                )}

                {/* Deny Inline Form */}
                {showDeny && (
                  <form onSubmit={handleDeny} className="bg-red-50 p-4 rounded-xl border border-red-200 animate-in slide-in-from-bottom-2">
                    <h4 className="font-bold text-red-900 mb-3 flex items-center">
                      <XCircle className="w-4 h-4 mr-2" /> Refuser la demande
                    </h4>
                    <div className="space-y-3">
                      <textarea placeholder="Motif clinique du refus (obligatoire, min 10 car.)" required minLength={10} rows={2} value={actionData.reason} onChange={e => setActionData({ ...actionData, reason: e.target.value })} className="w-full rounded border-red-300 p-2 text-sm focus:border-red-500 focus:ring-red-500"></textarea>
                      <div className="flex justify-end gap-2 pt-1">
                        <button type="button" onClick={() => setShowDeny(false)} className="px-4 py-2 text-red-700 bg-white border border-red-300 rounded font-medium hover:bg-red-100 text-sm">Annuler</button>
                        <button type="submit" disabled={actionLoading} className="px-6 py-2 bg-red-600 text-white rounded font-bold hover:bg-red-700 text-sm shadow-sm disabled:opacity-50">Confirmer le Refus</button>
                      </div>
                    </div>
                  </form>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {/* Image Preview Modal */}
      {previewImage && (
        <div className="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4" onClick={() => setPreviewImage(null)}>
          <div className="relative max-w-4xl max-h-[90vh] w-full">
            <button
              onClick={() => setPreviewImage(null)}
              className="absolute -top-10 right-0 text-white hover:text-gray-300 p-2"
            >
              <X className="w-8 h-8" />
            </button>
            <img
              src={previewImage.url}
              alt={previewImage.fileName}
              className="max-w-full max-h-[85vh] object-contain mx-auto rounded-lg shadow-2xl"
              onClick={(e) => e.stopPropagation()}
            />
            <div className="text-center text-white mt-3">
              <p className="font-bold text-sm">{previewImage.fileName}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
