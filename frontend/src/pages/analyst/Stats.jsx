import { useState, useEffect } from 'react';
import { getAnalystStats, getAnalystDoctorStats } from '../../services/api';
import {
  BarChart3,
  TrendingUp,
  Users,
  Activity,
  PieChart,
  Clock,
  AlertCircle,
  Building2,
  ChevronRight,
  UserCheck
} from 'lucide-react';

export default function AnalystDashboard() {
  const [stats, setStats] = useState([]);
  const [docStats, setDocStats] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const [deptData, doctorData] = await Promise.all([
        getAnalystStats(),
        getAnalystDoctorStats()
      ]);
      setStats(deptData || []);
      setDocStats(doctorData || []);
    } catch (err) {
      setError("Erreur lors de la récupération des données analytiques.");
    } finally {
      setLoading(false);
    }
  };

  if (loading) return (
    <div className="flex-1 flex justify-center items-center h-96">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand-600"></div>
    </div>
  );

  const totalReferrals = stats.reduce((acc, curr) => acc + curr.total_referrals, 0);
  const totalScheduled = stats.reduce((acc, curr) => acc + curr.scheduled_referrals, 0);
  const totalCritical = stats.reduce((acc, curr) => acc + curr.critical_urgency, 0);

  return (
    <div className="space-y-8 pb-12">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">Tableau de Bord Analyste</h1>
          <p className="text-gray-500 mt-1 font-medium">Analyse et indicateurs de performance par département CHU.</p>
        </div>
        <div className="flex items-center gap-2 bg-white px-4 py-2 rounded-xl shadow-sm border border-gray-100">
          <Clock className="w-4 h-4 text-brand-500" />
          <span className="text-sm font-bold text-gray-700">Dernière mise à jour: {new Date().toLocaleTimeString()}</span>
        </div>
      </div>

      {/* High-Level Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <StatCard
          icon={<Activity className="w-6 h-6" />}
          label="Total Demandes"
          value={totalReferrals}
          color="bg-brand-50 text-brand-700"
          trend="+12% ce mois"
        />
        <StatCard
          icon={<TrendingUp className="w-6 h-6" />}
          label="Taux d'Acceptation"
          value={totalReferrals > 0 ? `${Math.round((totalScheduled / totalReferrals) * 100)}%` : '0%'}
          color="bg-emerald-50 text-emerald-700"
          trend="En hausse"
        />
        <StatCard
          icon={<AlertCircle className="w-6 h-6" />}
          label="Cas Critiques"
          value={totalCritical}
          color="bg-red-50 text-red-700"
          trend="Nécessite attention"
        />
        <StatCard
          icon={<Users className="w-6 h-6" />}
          label="Départements Actifs"
          value={stats.filter(s => s.is_accepting).length}
          color="bg-blue-50 text-blue-700"
          trend="Opérationnels"
        />
      </div>

      {/* main Analytics Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

        {/* Department Volume Chart (Visual representation using Tailwind) */}
        <div className="lg:col-span-2 bg-white rounded-3xl p-8 shadow-sm border border-gray-100">
          <div className="flex items-center justify-between mb-8">
            <h2 className="text-xl font-bold text-gray-900 flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-brand-600" />
              Volume de Références par Département
            </h2>
          </div>

          <div className="space-y-6">
            {stats.sort((a,b) => b.total_referrals - a.total_referrals).map((dept) => {
              const percentage = totalReferrals > 0 ? (dept.total_referrals / stats.reduce((max, d) => Math.max(max, d.total_referrals), 0)) * 100 : 0;
              return (
                <div key={dept.id} className="group">
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-sm font-bold text-gray-700 group-hover:text-brand-600 transition-colors uppercase tracking-tight">{dept.name}</span>
                    <span className="text-sm font-black text-gray-900">{dept.total_referrals}</span>
                  </div>
                  <div className="h-4 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-brand-500 to-brand-400 transition-all duration-1000 ease-out"
                      style={{ width: `${percentage}%` }}
                    ></div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Breakdown Panel */}
        <div className="space-y-6">
          <div className="bg-gray-900 rounded-3xl p-8 text-white shadow-xl shadow-gray-200">
            <h3 className="text-lg font-bold mb-6 flex items-center gap-2">
              <PieChart className="w-5 h-5 text-brand-400" />
              Répartition par Urgence
            </h3>
            <div className="space-y-4">
              <UrgencyProgress label="CRITICAL" value={totalCritical} total={totalReferrals} color="bg-red-500" />
              <UrgencyProgress label="HIGH" value={stats.reduce((acc, c) => acc + c.high_urgency, 0)} total={totalReferrals} color="bg-orange-500" />
              <UrgencyProgress label="MEDIUM" value={stats.reduce((acc, c) => acc + c.medium_urgency, 0)} total={totalReferrals} color="bg-yellow-500" />
              <UrgencyProgress label="LOW" value={stats.reduce((acc, c) => acc + c.low_urgency, 0)} total={totalReferrals} color="bg-emerald-500" />
            </div>
          </div>

          <div className="bg-white rounded-3xl p-6 border border-gray-100 shadow-sm">
            <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
              <Users className="w-5 h-5 text-purple-600" />
              Top Doctors (CHU)
            </h3>
            <div className="space-y-3">
              {stats.slice(0, 3).map(dept => (
                <div key={dept.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-white border border-gray-200 rounded-full flex items-center justify-center font-bold text-xs text-brand-600">
                      {dept.doctors?.length || 0}
                    </div>
                    <span className="text-sm font-semibold text-gray-700">{dept.name}</span>
                  </div>
                  <ChevronRight className="w-4 h-4 text-gray-300" />
                </div>
              ))}
            </div>
          </div>
        </div>

      </div>

      {/* Doctor Referrals Section */}
      <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-50 flex items-center justify-between bg-gray-50/50">
          <div>
            <h2 className="text-xl font-bold text-gray-900">Activité des Médecins Référents (Niveau 2)</h2>
            <p className="text-xs text-gray-500 font-medium">Analyse des flux de patients par médecin et établissement d'origine.</p>
          </div>
          <Users className="w-6 h-6 text-brand-600" />
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-gray-50/30 border-b border-gray-100">
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest">Médecin / Établissement</th>
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest text-center">Total Référés</th>
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest">Répartition par Service</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {docStats.map(doc => (
                <tr key={doc.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-5">
                    <div className="font-bold text-gray-900 truncate max-w-[200px]">Dr. {doc.username}</div>
                    <div className="text-xs text-brand-600 font-bold bg-brand-50 inline-block px-1.5 py-0.5 rounded mt-1">{doc.facility_name}</div>
                  </td>
                  <td className="px-6 py-5 text-center font-black text-gray-900 text-lg">
                    {doc.total_referrals}
                  </td>
                  <td className="px-6 py-5">
                    <div className="flex flex-wrap gap-2 text-[10px]">
                      {doc.by_department && doc.by_department.length > 0 ? (
                        doc.by_department.map((dest, idx) => (
                          <div key={idx} className="flex items-center gap-1 bg-white border border-gray-200 rounded-full px-2 py-1 shadow-sm">
                            <span className="font-bold text-gray-700">{dest.name}:</span>
                            <span className="font-black text-brand-600">{dest.count}</span>
                          </div>
                        ))
                      ) : (
                        <span className="text-gray-400 italic">Aucune donnée</span>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Detail Table */}
      <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-50 flex items-center justify-between">
          <h2 className="text-xl font-bold text-gray-900">Statistiques Détaillées par Service</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-100">
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest">Département</th>
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest text-center">Volume Total</th>
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest text-center">En Attente</th>
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest text-center">Planifiés</th>
                <th className="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-widest text-center">Indice Critique</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {stats.map(dept => (
                <tr key={dept.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 bg-brand-50 rounded-xl flex items-center justify-center text-brand-600 font-bold">
                        {dept.name.charAt(0)}
                      </div>
                      <div>
                        <div className="font-bold text-gray-900">{dept.name}</div>
                        <div className="text-xs text-gray-500 font-medium">Poste: {dept.phone_extension || 'N/A'}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center font-bold text-gray-900">{dept.total_referrals}</td>
                  <td className="px-6 py-4 text-center">
                    <span className="px-3 py-1 bg-yellow-100 text-yellow-800 rounded-full text-xs font-bold">{dept.pending_referrals}</span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="px-3 py-1 bg-emerald-100 text-emerald-800 rounded-full text-xs font-bold">{dept.scheduled_referrals}</span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <div className="text-sm font-bold text-red-600">{dept.critical_urgency} cases</div>
                    <div className="flex justify-center mt-1">
                      {[1,2,3,4,5].map(i => (
                        <div key={i} className={`w-1 h-3 rounded-full mx-0.5 ${i <= (dept.critical_urgency/2) ? 'bg-red-500' : 'bg-gray-200'}`}></div>
                      ))}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function StatCard({ icon, label, value, color, trend }) {
  return (
    <div className="bg-white rounded-3xl p-6 border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
      <div className={`w-12 h-12 rounded-2xl ${color} flex items-center justify-center mb-4`}>
        {icon}
      </div>
      <div className="text-3xl font-black text-gray-900 mb-1 tracking-tight">{value}</div>
      <div className="text-sm font-bold text-gray-500 uppercase tracking-tight">{label}</div>
      <div className="mt-4 pt-4 border-t border-gray-50 text-[10px] font-black text-brand-500 uppercase tracking-widest">{trend}</div>
    </div>
  );
}

function UrgencyProgress({ label, value, total, color }) {
  const percentage = total > 0 ? (value / total) * 100 : 0;
  return (
    <div>
      <div className="flex justify-between items-center text-xs font-bold mb-1.5 uppercase tracking-widest opacity-80">
        <span>{label}</span>
        <span>{Math.round(percentage)}%</span>
      </div>
      <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
        <div 
          className={`h-full ${color} transition-all duration-1000`}
          style={{ width: `${percentage}%` }}
        ></div>
      </div>
    </div>
  );
}
