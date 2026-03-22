import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { checkNotifications } from '../../services/api';
import { subscribe as wsSubscribe, WS_EVENTS } from '../../services/websocket';
import { FilePlus, Bell, ArrowRight, XCircle, CheckCircle, ArrowRightLeft } from 'lucide-react';

export default function Level2Dashboard() {
  const { user } = useAuth();
  const [unreadCount, setUnreadCount] = useState(0);
  const [lastNotification, setLastNotification] = useState(null);

  useEffect(() => {
    // Quick polling or initial fetch for unread notifications
    const fetchAlerts = async () => {
      try {
        const data = await checkNotifications();
        setUnreadCount(data.unread_count || 0);
      } catch (err) {
        console.error('Failed to fetch notifications', err);
      }
    };
    fetchAlerts();

    // Subscribe to WebSocket notifications for referral status updates
    const unsubscribe = wsSubscribe((message) => {
      if (message.type === WS_EVENTS.REFERRAL_UPDATE) {
        // Referral status changed - refresh notification count
        fetchAlerts();
        // Show in-app notification
        let notification = {
          type: message.status,
          message: ''
        };

        if (message.status === 'SCHEDULED') {
          notification.message = `Votre demande #{message.referral_id} a été planifiée!`;
        } else if (message.status === 'DENIED') {
          notification.message = `Votre demande #{message.referral_id} a été refusée.`;
        } else if (message.status === 'REDIRECTED') {
          notification.message = `Votre demande #{message.referral_id} a été réorientée.`;
        }

        setLastNotification(notification);
        setTimeout(() => setLastNotification(null), 5000);
      } else if (message.type === WS_EVENTS.REFERRAL_REDIRECT) {
        // Referral was redirected - refresh count
        fetchAlerts();
        setLastNotification({
          type: 'REDIRECTED',
          message: `Votre demande #{message.referral_id} a été réorientée vers un autre service.`
        });
        setTimeout(() => setLastNotification(null), 5000);
      }
    });

    return () => unsubscribe();
  }, []);

  return (
    <div className="space-y-6">

      {/* Real-time Notification Toast */}
      {lastNotification && (
        <div className="fixed top-20 right-4 z-50 animate-in slide-in-from-right-2 fade-in duration-300">
          <div className={`px-5 py-4 rounded-xl shadow-lg border max-w-sm ${lastNotification.type === 'SCHEDULED' ? 'bg-green-50 border-green-200' :
              lastNotification.type === 'DENIED' ? 'bg-red-50 border-red-200' :
                'bg-blue-50 border-blue-200'
            }`}>
            <div className="flex items-start gap-3">
              <div className={`p-2 rounded-lg ${lastNotification.type === 'SCHEDULED' ? 'bg-green-100' :
                  lastNotification.type === 'DENIED' ? 'bg-red-100' :
                    'bg-blue-100'
                }`}>
                {lastNotification.type === 'SCHEDULED' && <CheckCircle className="w-5 h-5 text-green-600" />}
                {lastNotification.type === 'DENIED' && <XCircle className="w-5 h-5 text-red-600" />}
                {lastNotification.type === 'REDIRECTED' && <ArrowRightLeft className="w-5 h-5 text-blue-600" />}
              </div>
              <div className="flex-1">
                <h4 className="font-bold text-sm text-gray-900">Mise à jour de votre demande</h4>
                <p className="text-xs text-gray-600 mt-1">{lastNotification.message}</p>
              </div>
              <button
                onClick={() => setLastNotification(null)}
                className="text-gray-400 hover:text-gray-600"
              >
                <XCircle className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Welcome Banner */}
      <div className="bg-brand-600 rounded-2xl p-6 text-white shadow-md relative overflow-hidden">
        <div className="absolute top-0 right-0 w-64 h-64 bg-white opacity-5 rounded-full -translate-y-1/2 translate-x-1/2"></div>
        <h1 className="text-2xl font-bold relative z-10">Bonjour, Dr. {user.username}</h1>
        <p className="text-brand-100 mt-1 relative z-10">{user.facility_name}</p>
      </div>

      {/* Quick Actions */}
      <h2 className="text-lg font-bold text-gray-900 px-2 mt-8">Actions Rapides</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">

        <Link
          to="/referrals/new"
          className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 flex items-center hover:border-brand-300 hover:shadow-md transition-all group"
        >
          <div className="bg-brand-50 rounded-xl p-4 mr-4 group-hover:bg-brand-100 transition-colors">
            <FilePlus className="w-8 h-8 text-brand-600" />
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-bold text-gray-900">Nouvelle Référence</h3>
            <p className="text-sm text-gray-500">Créer un dossier patient pour le CHU</p>
          </div>
          <ArrowRight className="w-5 h-5 text-gray-300 group-hover:text-brand-500 transform group-hover:translate-x-1 transition-all" />
        </Link>

        <Link
          to="/notifications"
          className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 flex items-center hover:border-brand-300 hover:shadow-md transition-all group"
        >
          <div className="bg-blue-50 rounded-xl p-4 mr-4 group-hover:bg-blue-100 transition-colors relative">
            <Bell className="w-8 h-8 text-blue-600" />
            {unreadCount > 0 && (
              <span className="absolute top-2 right-2 flex h-3 w-3">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>
              </span>
            )}
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-bold text-gray-900">Notifications</h3>
            <p className="text-sm text-gray-500">
              {unreadCount > 0
                ? <span className="text-red-600 font-medium">{unreadCount} non lues</span>
                : "Aucune nouvelle alerte"}
            </p>
          </div>
          <ArrowRight className="w-5 h-5 text-gray-300 group-hover:text-blue-500 transform group-hover:translate-x-1 transition-all" />
        </Link>
      </div>

      {/* Info Card */}
      <div className="bg-amber-50 rounded-2xl p-5 border border-amber-200 mt-8">
        <h3 className="text-sm font-bold text-amber-800 mb-1">Rappel de Sécurité</h3>
        <p className="text-xs text-amber-700 leading-relaxed">
          Toutes les données patients sont chiffrées de bout en bout. L'accès est tracé et réservé à l'usage strictement médical selon la loi 09-08 (CNDP).
        </p>
      </div>

    </div>
  );
}
