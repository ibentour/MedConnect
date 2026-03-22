import { useState, useEffect } from 'react';
import { checkNotifications, markNotificationRead } from '../../services/api';
import { Bell, CheckCircle2, AlertTriangle, ArrowRightLeft, XCircle } from 'lucide-react';

export default function Notifications() {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadNotifications();
  }, []);

  const loadNotifications = async () => {
    try {
      const data = await checkNotifications();
      setNotifications(data.notifications || []);
    } catch (error) {
      console.error('Failed to load notifications', error);
    } finally {
      setLoading(false);
    }
  };

  const handleMarkRead = async (id, e) => {
    e.preventDefault();
    try {
      await markNotificationRead(id);
      // Optimistic UI update
      setNotifications(notifications.map(n => 
        n.id === id ? { ...n, is_read: true } : n
      ));
    } catch (error) {
      console.error('Failed to mark read', error);
    }
  };

  const getIcon = (message) => {
    if (message.includes('✅')) return <CheckCircle2 className="w-6 h-6 text-green-500" />;
    if (message.includes('❌')) return <XCircle className="w-6 h-6 text-red-500" />;
    if (message.includes('🔄')) return <ArrowRightLeft className="w-6 h-6 text-blue-500" />;
    return <Bell className="w-6 h-6 text-gray-500" />;
  };

  return (
    <div className="space-y-6 max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Notifications</h1>
      </div>

      {loading ? (
        <div className="flex justify-center p-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-600"></div>
        </div>
      ) : notifications.length === 0 ? (
        <div className="bg-white p-12 rounded-2xl border border-gray-100 text-center">
          <Bell className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900">Aucune notification</h3>
          <p className="text-gray-500 mt-1">Vous serez alerté ici des mises à jour de vos dossiers.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {notifications.map((notif) => (
            <div 
              key={notif.id} 
              className={`p-4 rounded-xl border transition-all flex gap-4 ${
                notif.is_read 
                  ? 'bg-white border-gray-100 opacity-75' 
                  : 'bg-brand-50 border-brand-200 shadow-sm'
              }`}
            >
              <div className="flex-shrink-0 mt-1">
                {getIcon(notif.message)}
              </div>
              <div className="flex-1">
                <p className={`text-sm md:text-base leading-relaxed ${notif.is_read ? 'text-gray-700' : 'text-gray-900 font-medium'}`}>
                  {/* Remove the emoji from the text since we have an icon */}
                  {notif.message.replace(/^[✅❌🔄]\s*/, '')}
                </p>
                <div className="mt-2 text-xs text-gray-500 flex items-center justify-between">
                  <span>{new Date(notif.created_at).toLocaleString('fr-FR')}</span>
                  {!notif.is_read && (
                    <button 
                      onClick={(e) => handleMarkRead(notif.id, e)}
                      className="text-brand-600 hover:text-brand-800 font-medium px-2 py-1 rounded hover:bg-brand-100 transition-colors"
                    >
                      Marquer comme lu
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
