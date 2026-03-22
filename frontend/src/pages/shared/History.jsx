import { useState, useEffect } from 'react';
import { getHistory, rescheduleReferral, cancelReferral, downloadAttachment } from '../../services/api';
import { useAuth } from '../../context/AuthContext';
import { StatusBadge } from '../../components/StatusBadge';
import { Calendar, XCircle, Clock, Building2, UserCircle2, CheckCircle2, Copy, Paperclip, File as FileIcon, ExternalLink } from 'lucide-react';

export default function HistoryTab() {
  const { user } = useAuth();
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(true);

  // Rescheduling state
  const [rescheduleData, setRescheduleData] = useState({ id: null, date: '' });

  useEffect(() => {
    loadHistory();
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

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 border-b-4 border-blue-600 pb-1">
          {user?.role === 'LEVEL_2_DOC' ? 'Historique des envois' : 'Historique des dossiers'}
        </h1>
        <div className="bg-white px-4 py-2 rounded-full shadow-sm text-sm font-medium text-gray-500 border border-gray-200">
          Total: <span className="font-bold text-gray-900">{history.length}</span> enregistrements
        </div>
      </div>

      {loading ? (
        <div className="text-center p-12 text-gray-500 font-medium animate-pulse">Chargement de l'historique...</div>
      ) : history.length === 0 ? (
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-12 text-center">
          <Clock className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-bold text-gray-900">Aucun historique</h3>
          <p className="text-gray-500">Les dossiers clôs apparaîtront ici.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {history.map((item) => (
            <div key={item.id} className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow flex flex-col">
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
    </div>
  );
}
