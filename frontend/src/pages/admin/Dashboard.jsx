import React, { useState, useEffect, Fragment } from 'react';

// Time Picker Component - intuitive hour and minute selection
function TimePicker({ value, onChange, label = 'De' }) {
  // Parse the current value (expected format: "HH:MM-HH:MM")
  const parseTime = (timeStr) => {
    if (!timeStr || !timeStr.includes('-')) return { start: '08:00', end: '16:00' };
    const [start, end] = timeStr.split('-');
    return { start: start.trim(), end: end.trim() };
  };

  const { start, end } = parseTime(value);
  const [startHour, startMin] = start.split(':');
  const [endHour, endMin] = end.split(':');

  const hours = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0'));
  const minutes = ['00', '15', '30', '45'];

  const handleStartChange = (h, m) => {
    onChange(`${h}:${m}-${endHour}:${endMin}`);
  };

  const handleEndChange = (h, m) => {
    onChange(`${startHour}:${startMin}-${h}:${m}`);
  };

  return (
    <div className="flex flex-col gap-2 p-2 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden max-w-full">
      <div className="flex items-center gap-2">
        <span className="text-xs text-gray-500 font-medium w-8">De:</span>
        <div className="flex items-center gap-1">
          <select
            value={startHour}
            onChange={(e) => handleStartChange(e.target.value, startMin)}
            className="border border-gray-200 rounded-lg p-2 text-sm bg-white focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all"
          >
            {hours.map(h => <option key={`start-h-${h}`} value={h}>{h}</option>)}
          </select>
          <span className="text-gray-400 font-bold">:</span>
          <select
            value={startMin}
            onChange={(e) => handleStartChange(startHour, e.target.value)}
            className="border border-gray-200 rounded-lg p-2 text-sm bg-white w-16 focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all"
          >
            {minutes.map(m => <option key={`start-m-${m}`} value={m}>{m}</option>)}
          </select>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <span className="text-xs text-gray-500 font-medium w-8">À:</span>
        <div className="flex items-center gap-1">
          <select
            value={endHour}
            onChange={(e) => handleEndChange(e.target.value, endMin)}
            className="border border-gray-200 rounded-lg p-2 text-sm bg-white focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all"
          >
            {hours.map(h => <option key={`end-h-${h}`} value={h}>{h}</option>)}
          </select>
          <span className="text-gray-400 font-bold">:</span>
          <select
            value={endMin}
            onChange={(e) => handleEndChange(endHour, e.target.value)}
            className="border border-gray-200 rounded-lg p-2 text-sm bg-white w-16 focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all"
          >
            {minutes.map(m => <option key={`end-m-${m}`} value={m}>{m}</option>)}
          </select>
        </div>
      </div>
    </div>
  );
}
import { getAdminStats, getUsers, createUser, deleteUser, getAdminDepartments, createDepartment, deleteDepartment, updateDepartment } from '../../services/api';
import { AntiLeak } from '../../components/Security';
import { Users, Building2, Activity, Trash2, Plus, RefreshCw, ChevronDown, ChevronRight, Save, UserCircle2, Clock, Phone, Settings } from 'lucide-react';

export default function AdminDashboard() {
  const [stats, setStats] = useState(null);
  const [users, setUsers] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [loading, setLoading] = useState(true);

  const [expandedDept, setExpandedDept] = useState(null);
  const [editDept, setEditDept] = useState(null); // holds the dept currently being edited

  const DAYS = ['Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam', 'Dim'];

  const toggleDay = (currentStr, day) => {
    const selected = currentStr ? currentStr.split(',').map(s => s.trim()).filter(Boolean) : [];
    if (selected.includes(day)) {
      return selected.filter(d => d !== day).join(',');
    } else {
      const newSelected = new Set([...selected, day]);
      return DAYS.filter(d => newSelected.has(d)).join(',');
    }
  };

  // Forms
  const [newUser, setNewUser] = useState({ username: '', password: '', role: 'LEVEL_2_DOC', facility_name: '', department_id: '' });
  const [newDept, setNewDept] = useState({ name: '', phone_extension: '', work_hours: '', work_days: '' });

  useEffect(() => {
    loadAll();
  }, []);

  const loadAll = async () => {
    setLoading(true);
    try {
      const [statsData, usersData, deptsData] = await Promise.all([
        getAdminStats(),
        getUsers(),
        getAdminDepartments()
      ]);
      setStats(statsData);
      setUsers(usersData);
      setDepartments(deptsData);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateUser = async (e) => {
    e.preventDefault();
    try {
      const payload = { ...newUser };
      if (payload.role !== 'CHU_DOC') delete payload.department_id;
      await createUser(payload);
      setNewUser({ username: '', password: '', role: 'LEVEL_2_DOC', facility_name: '', department_id: '' });
      loadAll();
    } catch (err) {
      alert("Erreur: " + (err.response?.data?.error || err.message));
    }
  };

  const handleDeleteUser = async (id) => {
    if (!window.confirm("Delete this user?")) return;
    try {
      await deleteUser(id);
      loadAll();
    } catch (err) {
      alert("Erreur: " + err.message);
    }
  };

  const handleCreateDept = async (e) => {
    e.preventDefault();
    try {
      await createDepartment(newDept);
      setNewDept({ name: '', phone_extension: '', work_hours: '', work_days: '' });
      loadAll();
    } catch (err) {
      alert("Erreur: " + (err.response?.data?.error || err.message));
    }
  };

  const handleDeleteDept = async (id) => {
    if (!window.confirm("Delete this department?")) return;
    try {
      await deleteDepartment(id);
      loadAll();
    } catch (err) {
      alert("Erreur: " + err.message);
    }
  };

  const handleUpdateDept = async (e) => {
    e.preventDefault();
    try {
      await updateDepartment(editDept.id, {
        name: editDept.name,
        phone_extension: editDept.phone_extension,
        work_hours: editDept.work_hours,
        work_days: editDept.work_days
      });
      setEditDept(null);
      loadAll();
    } catch (err) {
      alert("Erreur: " + (err.response?.data?.error || err.message));
    }
  };

  const toggleExpand = (d) => {
    if (expandedDept === d.id) {
      setExpandedDept(null);
      setEditDept(null);
    } else {
      setExpandedDept(d.id);
      setEditDept({ ...d });
    }
  };

  if (loading && !stats) return (
    <div className="flex-1 flex justify-center items-center h-96">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand-600"></div>
    </div>
  );

  return (
    <div className="space-y-8 pb-12">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">Tableau de Bord Administrateur</h1>
          <p className="text-gray-500 mt-1 font-medium">Gestion des utilisateurs, services et supervision du système.</p>
        </div>
      </div>

      {/* STATS */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-brand-50 p-4 rounded-2xl text-brand-600"><Users className="w-6 h-6" /></div>
          <div><p className="text-3xl font-black text-gray-900">{stats.total_users}</p><p className="text-xs text-gray-500 font-bold tracking-wide">Utilisateurs</p></div>
        </div>
        <div className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-brand-50 p-4 rounded-2xl text-brand-600"><Building2 className="w-6 h-6" /></div>
          <div><p className="text-3xl font-black text-gray-900">{stats.total_departments}</p><p className="text-xs text-gray-500 font-bold tracking-wide">Services CHU</p></div>
        </div>
        <div className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-emerald-50 p-4 rounded-2xl text-emerald-600"><Activity className="w-6 h-6" /></div>
          <div><p className="text-3xl font-black text-gray-900">{stats.total_referrals}</p><p className="text-xs text-gray-500 font-bold tracking-wide">Dossiers Total</p></div>
        </div>
        <div className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-amber-50 p-4 rounded-2xl text-amber-600"><RefreshCw className="w-6 h-6" /></div>
          <div><p className="text-3xl font-black text-gray-900">{stats.pending_referrals}</p><p className="text-xs text-gray-500 font-bold tracking-wide">En Attente</p></div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">

        {/* USERS MANAGEMENT */}
        <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
          <div className="px-6 py-5 border-b border-gray-50 bg-gray-50/50">
            <h2 className="text-xl font-bold text-gray-900 flex items-center gap-2">
              <Users className="w-5 h-5 text-brand-600" />
              Gestion des Médecins / Utilisateurs
            </h2>
          </div>
          <div className="p-5 overflow-auto max-h-[400px]">
            <table className="w-full text-sm text-left">
              <thead className="text-xs font-bold text-gray-400 uppercase tracking-widest bg-gray-50/30">
                <tr><th className="px-4 py-3">Username</th><th className="px-4 py-3">Rôle</th><th className="px-4 py-3">Structure</th><th className="px-4 py-3">Actions</th></tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {users.map(u => (
                  <tr key={u.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-4 font-bold text-gray-900">{u.username}</td>
                    <td className="px-4 py-4"><span className="bg-brand-50 text-brand-700 px-2.5 py-1 rounded-full text-xs font-bold">{u.role === 'LEVEL_2_DOC' ? 'Level 2' : u.role === 'CHU_DOC' ? 'CHU' : u.role}</span></td>
                    <td className="px-4 py-4 text-gray-500">{u.facility_name} {u.department ? `(${u.department.name})` : ''}</td>
                    <td className="px-4 py-4"><button onClick={() => handleDeleteUser(u.id)} className="text-red-500 hover:text-red-700 p-2 hover:bg-red-50 rounded-lg transition-colors"><Trash2 className="w-4 h-4" /></button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <form onSubmit={handleCreateUser} className="p-6 border-t border-gray-100 bg-gray-50/30 grid grid-cols-1 md:grid-cols-2 gap-4">
            <input type="text" placeholder="Username" required value={newUser.username} onChange={e => setNewUser({ ...newUser, username: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all" />
            <input type="password" placeholder="Mot de passe (min 8)" required value={newUser.password} onChange={e => setNewUser({ ...newUser, password: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all" />
            <select value={newUser.role} onChange={e => setNewUser({ ...newUser, role: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all">
              <option value="LEVEL_2_DOC">Level 2 Doctor</option>
              <option value="CHU_DOC">CHU Doctor</option>
              <option value="ANALYST">Analyst</option>
              <option value="SUPER_ADMIN">Super Admin</option>
            </select>
            <input type="text" placeholder="Établissement" required value={newUser.facility_name} onChange={e => setNewUser({ ...newUser, facility_name: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all" />
            {newUser.role === 'CHU_DOC' && (
              <select required value={newUser.department_id} onChange={e => setNewUser({ ...newUser, department_id: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all md:col-span-2">
                <option value="">Sélectionner service CHU...</option>
                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            )}
            <button type="submit" className="md:col-span-2 bg-brand-600 text-white font-bold py-3 px-6 rounded-xl flex justify-center items-center hover:bg-brand-700 transition-colors shadow-sm"><Plus className="w-4 h-4 mr-2" /> Ajouter un Utilisateur</button>
          </form>
        </div>

        {/* DEPARTMENTS MANAGEMENT */}
        <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
          <div className="px-6 py-5 border-b border-gray-50 bg-gray-50/50">
            <h2 className="text-xl font-bold text-gray-900 flex items-center gap-2">
              <Building2 className="w-5 h-5 text-brand-600" />
              Services Cliniques CHU
            </h2>
          </div>
          <div className="p-5 overflow-auto max-h-[400px]">
            <table className="w-full text-sm text-left">
              <thead className="text-xs font-bold text-gray-400 uppercase tracking-widest bg-gray-50/30">
                <tr><th className="px-4 py-3">Service</th><th className="px-4 py-3 text-center">En attente</th><th className="px-4 py-3 text-center">Planifiés</th><th className="px-4 py-3">Actions</th></tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {departments.map(d => (
                  <React.Fragment key={d.id}>
                    <tr className={`cursor-pointer hover:bg-gray-50 transition-colors ${expandedDept === d.id ? 'bg-brand-50/30' : ''}`} onClick={() => toggleExpand(d)}>
                      <td className="px-4 py-4 flex items-center gap-2">
                        {expandedDept === d.id ? <ChevronDown className="w-4 h-4 text-brand-600" /> : <ChevronRight className="w-4 h-4 text-gray-400" />}
                        <span className="font-bold text-gray-900">{d.name}</span>
                      </td>
                      <td className="px-4 py-4 text-center"><span className="text-amber-600 font-black">{d.pending_referrals}</span></td>
                      <td className="px-4 py-4 text-center"><span className="text-emerald-600 font-black">{d.scheduled_referrals}</span></td>
                      <td className="px-4 py-4" onClick={e => e.stopPropagation()}>
                        <button onClick={() => handleDeleteDept(d.id)} className="text-red-500 hover:text-red-700 p-2 hover:bg-red-50 rounded-lg transition-colors"><Trash2 className="w-4 h-4" /></button>
                      </td>
                    </tr>

                    {/* EXPANDED ANALYTICS & EDIT FORM */}
                    {expandedDept === d.id && (
                      <tr>
                        <td colSpan="4" className="bg-brand-50/20 p-6 border-b border-brand-100">
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">

                            {/* Analytics Panel */}
                            <div className="space-y-4">
                              <h4 className="text-xs uppercase font-bold text-brand-700 tracking-wider flex items-center gap-2">
                                <Activity className="w-4 h-4" /> Analytics & Urgences
                              </h4>
                              <div className="flex flex-wrap gap-3 justify-center">
                                <div className="flex-1 min-w-[80px] bg-white border border-gray-100 p-3 rounded-2xl text-center shadow-sm">
                                  <div className="text-[10px] text-gray-400 uppercase font-bold mb-1 tracking-wide">Low</div>
                                  <div className="text-2xl font-black text-blue-600">{d.low_urgency}</div>
                                </div>
                                <div className="flex-1 min-w-[80px] bg-white border border-gray-100 p-3 rounded-2xl text-center shadow-sm">
                                  <div className="text-[10px] text-gray-400 uppercase font-bold mb-1 tracking-wide">Medium</div>
                                  <div className="text-2xl font-black text-yellow-600">{d.medium_urgency}</div>
                                </div>
                                <div className="flex-1 min-w-[80px] bg-white border border-gray-100 p-3 rounded-2xl text-center shadow-sm">
                                  <div className="text-[10px] text-gray-400 uppercase font-bold mb-1 tracking-wide">High</div>
                                  <div className="text-2xl font-black text-orange-600">{d.high_urgency}</div>
                                </div>
                                <div className="flex-1 min-w-[80px] bg-white border border-gray-100 p-3 rounded-2xl text-center shadow-sm">
                                  <div className="text-[10px] text-gray-400 uppercase font-bold mb-1 tracking-wide">Critical</div>
                                  <div className="text-2xl font-black text-red-600">{d.critical_urgency}</div>
                                </div>
                              </div>

                              <h4 className="text-xs uppercase font-bold text-brand-700 tracking-wider flex items-center gap-2 mt-6">
                                <UserCircle2 className="w-4 h-4" /> Docteurs Assignés ({d.doctors?.length || 0})
                              </h4>
                              <div className="flex flex-wrap gap-2">
                                {d.doctors?.length > 0 ? d.doctors.map(doc => (
                                  <span key={doc.id} className="text-xs bg-brand-100 text-brand-700 px-3 py-1.5 rounded-full font-semibold">{doc.username}</span>
                                )) : <span className="text-xs text-gray-400 italic">Aucun docteur assigné</span>}
                              </div>
                            </div>

                            {/* Edit Form */}
                            <div className="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm">
                              <h4 className="text-xs uppercase font-bold text-gray-900 tracking-wider mb-4 flex items-center gap-2">
                                <Settings className="w-4 h-4" /> Modifier le Service
                              </h4>
                              <form onSubmit={handleUpdateDept} className="space-y-3">
                                <div>
                                  <label className="text-xs text-gray-500 mb-1.5 block font-medium">Nom du Service</label>
                                  <input type="text" value={editDept?.name || ''} onChange={e => setEditDept({ ...editDept, name: e.target.value })} className="w-full border border-gray-200 rounded-xl p-2.5 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all" required />
                                </div>
                                <div>
                                  <div className="mb-3">
                                    <label className="text-xs text-gray-500 mb-1.5 block font-medium flex items-center gap-1">
                                      <Phone className="w-3 h-3" /> Numéro De Téléphone
                                    </label>
                                    <input
                                      type="tel"
                                      value={editDept?.phone_extension || ''}
                                      onChange={e => setEditDept({ ...editDept, phone_extension: e.target.value })}
                                      placeholder="+212 6XX XXX XXX"
                                      pattern="[+]?[0-9\s\-\(\)]{10,}"
                                      className="w-full border border-gray-200 rounded-xl p-2.5 text-sm focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all"
                                    />
                                  </div>
                                  <div className="mb-3">
                                    <label className="text-xs text-gray-500 mb-1.5 block font-medium flex items-center gap-1">
                                      <Clock className="w-3 h-3" /> Heures d'Ouverture
                                    </label>
                                    <TimePicker
                                      value={editDept?.work_hours || ''}
                                      onChange={(hours) => setEditDept({ ...editDept, work_hours: hours })}
                                    />
                                  </div>
                                  <div>
                                    <label className="text-xs text-gray-500 mb-1.5 block font-medium">Jours de Travail</label>
                                    <div className="flex gap-1.5 flex-wrap">
                                      {DAYS.map(day => {
                                        const isSelected = editDept?.work_days?.split(',').includes(day);
                                        return (
                                          <button
                                            type="button"
                                            key={`edit-${day}`}
                                            onClick={() => setEditDept({ ...editDept, work_days: toggleDay(editDept?.work_days, day) })}
                                            className={`px-3 py-1.5 text-xs rounded-xl border transition-all ${isSelected ? 'bg-brand-600 border-brand-600 text-white font-bold shadow-sm' : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50 hover:border-gray-300'}`}
                                          >
                                            {day}
                                          </button>
                                        )
                                      })}
                                    </div>
                                  </div>
                                </div>
                                <button type="submit" className="w-full mt-3 bg-brand-600 text-white font-bold py-2.5 rounded-xl text-sm hover:bg-brand-700 flex justify-center items-center gap-3 transition-colors shadow-sm">
                                  Enregistrer les Modifications <Save className="w-4 h-4 ml-1" />
                                </button>
                              </form>
                            </div>

                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
          <form onSubmit={handleCreateDept} className="p-6 border-t border-gray-100 bg-gray-50/30 space-y-4">
            <h3 className="text-base font-bold text-gray-800 mb-4 flex items-center gap-2">
              <Plus className="w-5 h-5 text-brand-600" />
              Ajouter un Nouveau Service
            </h3>
            <input type="text" placeholder="Nom du service (ex: Radiologie)" required value={newDept.name} onChange={e => setNewDept({ ...newDept, name: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm w-full focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all" />

            <div>
              <label className="text-xs text-gray-500 mb-1.5 block font-medium"><Phone className="w-3 h-3 inline mr-1" />Numéro De Téléphone</label>
              <input type="tel" placeholder="+212 6XX XXX XXX" value={newDept.phone_extension} onChange={e => setNewDept({ ...newDept, phone_extension: e.target.value })} className="border border-gray-200 rounded-xl p-3 text-sm w-full focus:ring-2 focus:ring-brand-500 focus:border-brand-500 outline-none transition-all" pattern="[+]?[0-9\s\-\(\)]{10,}" />
            </div>

            <div>
              <label className="text-xs text-gray-500 mb-1.5 block font-medium">Heures d'Ouverture</label>
              <TimePicker
                value={newDept.work_hours || '08:00-16:00'}
                onChange={(hours) => setNewDept({ ...newDept, work_hours: hours })}
              />
            </div>

            <div>
              <label className="text-xs text-gray-500 mb-1.5 block font-medium">Jours de Travail</label>
              <div className="flex gap-2 flex-wrap">
                {DAYS.map(day => {
                  const isSelected = newDept.work_days?.split(',').includes(day);
                  return (
                    <button
                      type="button"
                      key={`new-${day}`}
                      onClick={() => setNewDept({ ...newDept, work_days: toggleDay(newDept.work_days, day) })}
                      className={`px-3 py-1.5 text-[11px] uppercase tracking-wider rounded-lg border transition-all ${isSelected ? 'bg-brand-600 border-brand-600 text-white font-bold shadow-sm' : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50 hover:border-gray-300'}`}
                    >
                      {day}
                    </button>
                  )
                })}
              </div>
            </div>

            <button type="submit" className="w-full bg-brand-600 text-white font-bold py-3 rounded-xl flex justify-center items-center gap-2 hover:bg-brand-700 transition-colors shadow-md">
              <Plus className="w-5 h-5" />
              Ajouter Service
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
