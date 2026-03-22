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
    <div className="flex items-center gap-1">
      <div className="flex items-center">
        <select
          value={startHour}
          onChange={(e) => handleStartChange(e.target.value, startMin)}
          className="border border-gray-300 rounded p-1 text-sm bg-white"
        >
          {hours.map(h => <option key={`start-h-${h}`} value={h}>{h}</option>)}
        </select>
        <span className="mx-0.5 text-gray-500">:</span>
        <select
          value={startMin}
          onChange={(e) => handleStartChange(startHour, e.target.value)}
          className="border border-gray-300 rounded p-1 text-sm bg-white w-14"
        >
          {minutes.map(m => <option key={`start-m-${m}`} value={m}>{m}</option>)}
        </select>
      </div>
      <span className="text-gray-400 mx-1">-</span>
      <div className="flex items-center">
        <select
          value={endHour}
          onChange={(e) => handleEndChange(e.target.value, endMin)}
          className="border border-gray-300 rounded p-1 text-sm bg-white"
        >
          {hours.map(h => <option key={`end-h-${h}`} value={h}>{h}</option>)}
        </select>
        <span className="mx-0.5 text-gray-500">:</span>
        <select
          value={endMin}
          onChange={(e) => handleEndChange(endHour, e.target.value)}
          className="border border-gray-300 rounded p-1 text-sm bg-white w-14"
        >
          {minutes.map(m => <option key={`end-m-${m}`} value={m}>{m}</option>)}
        </select>
      </div>
    </div>
  );
}
import { getAdminStats, getUsers, createUser, deleteUser, getAdminDepartments, createDepartment, deleteDepartment, updateDepartment } from '../../services/api';
import { AntiLeak } from '../../components/Security';
import { Users, Building2, Activity, Trash2, Plus, RefreshCw, ChevronDown, ChevronRight, Save, UserCircle2, Clock, Phone } from 'lucide-react';

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

  if (loading && !stats) return <div className="p-12 text-center text-gray-500">Chargement...</div>;

  return (
    <div className="space-y-8 max-w-7xl mx-auto">

      {/* STATS */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white p-5 rounded-xl shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-blue-100 p-3 rounded-lg text-blue-600"><Users className="w-6 h-6" /></div>
          <div><p className="text-2xl font-bold text-gray-900">{stats.total_users}</p><p className="text-xs text-gray-500 uppercase font-bold tracking-wider">Utilisateurs</p></div>
        </div>
        <div className="bg-white p-5 rounded-xl shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-indigo-100 p-3 rounded-lg text-indigo-600"><Building2 className="w-6 h-6" /></div>
          <div><p className="text-2xl font-bold text-gray-900">{stats.total_departments}</p><p className="text-xs text-gray-500 uppercase font-bold tracking-wider">Services CHU</p></div>
        </div>
        <div className="bg-white p-5 rounded-xl shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-green-100 p-3 rounded-lg text-green-600"><Activity className="w-6 h-6" /></div>
          <div><p className="text-2xl font-bold text-gray-900">{stats.total_referrals}</p><p className="text-xs text-gray-500 uppercase font-bold tracking-wider">Dossiers Total</p></div>
        </div>
        <div className="bg-white p-5 rounded-xl shadow-sm border border-gray-100 flex items-center gap-4">
          <div className="bg-amber-100 p-3 rounded-lg text-amber-600"><RefreshCw className="w-6 h-6" /></div>
          <div><p className="text-2xl font-bold text-gray-900">{stats.pending_referrals}</p><p className="text-xs text-gray-500 uppercase font-bold tracking-wider">En Attente</p></div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">

        {/* USERS MANAGEMENT */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-5 py-4 border-b border-gray-100 bg-gray-50 font-bold text-gray-900">Gestion des Médecins / Utilisateurs</div>
          <div className="p-5 overflow-auto max-h-[400px]">
            <table className="w-full text-sm text-left">
              <thead className="text-xs text-gray-500 uppercase bg-gray-50">
                <tr><th>Username</th><th>Rôle</th><th>Structure</th><th>Actions</th></tr>
              </thead>
              <tbody>
                {users.map(u => (
                  <tr key={u.id} className="border-b">
                    <td className="py-3 font-medium text-gray-900">{u.username}</td>
                    <td className="py-3"><span className="bg-blue-50 text-blue-700 px-2.5 py-0.5 rounded-full text-xs font-bold">{u.role}</span></td>
                    <td className="py-3 text-gray-500">{u.facility_name} {u.department ? `(${u.department.name})` : ''}</td>
                    <td className="py-3"><button onClick={() => handleDeleteUser(u.id)} className="text-red-500 hover:text-red-700"><Trash2 className="w-4 h-4" /></button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <form onSubmit={handleCreateUser} className="p-5 border-t border-gray-100 bg-gray-50 grid grid-cols-2 gap-3">
            <input type="text" placeholder="Username" required value={newUser.username} onChange={e => setNewUser({ ...newUser, username: e.target.value })} className="border rounded p-2 text-sm" />
            <input type="password" placeholder="Mot de passe (min 8)" required value={newUser.password} onChange={e => setNewUser({ ...newUser, password: e.target.value })} className="border rounded p-2 text-sm" />
            <select value={newUser.role} onChange={e => setNewUser({ ...newUser, role: e.target.value })} className="border rounded p-2 text-sm">
              <option value="LEVEL_2_DOC">LEVEL_2_DOC</option>
              <option value="CHU_DOC">CHU_DOC</option>
              <option value="ANALYST">ANALYST</option>
              <option value="SUPER_ADMIN">SUPER_ADMIN</option>
            </select>
            <input type="text" placeholder="Établissement" required value={newUser.facility_name} onChange={e => setNewUser({ ...newUser, facility_name: e.target.value })} className="border rounded p-2 text-sm" />
            {newUser.role === 'CHU_DOC' && (
              <select required value={newUser.department_id} onChange={e => setNewUser({ ...newUser, department_id: e.target.value })} className="border rounded p-2 text-sm col-span-2">
                <option value="">Sélectionner service CHU...</option>
                {departments.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
              </select>
            )}
            <button type="submit" className="col-span-2 bg-blue-600 text-white font-bold py-2 rounded flex justify-center items-center hover:bg-blue-700"><Plus className="w-4 h-4 mr-1" /> Ajouter</button>
          </form>
        </div>

        {/* DEPARTMENTS MANAGEMENT */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-5 py-4 border-b border-gray-100 bg-gray-50 font-bold text-gray-900">Services Cliniques CHU</div>
          <div className="p-5 overflow-auto max-h-[400px]">
            <table className="w-full text-sm text-left">
              <thead className="text-xs text-gray-500 uppercase bg-gray-50">
                <tr><th>Service</th><th>En attente</th><th>Planifiés</th><th>Actions</th></tr>
              </thead>
              <tbody>
                {departments.map(d => (
                  <React.Fragment key={d.id}>
                    <tr className={`border-b cursor-pointer hover:bg-gray-50 transition-colors ${expandedDept === d.id ? 'bg-indigo-50/50' : ''}`} onClick={() => toggleExpand(d)}>
                      <td className="py-3 px-2 flex items-center gap-2">
                        {expandedDept === d.id ? <ChevronDown className="w-4 h-4 text-indigo-500" /> : <ChevronRight className="w-4 h-4 text-gray-400" />}
                        <span className="font-bold text-gray-900">{d.name}</span>
                      </td>
                      <td className="py-3"><span className="text-amber-600 font-bold">{d.pending_referrals}</span></td>
                      <td className="py-3"><span className="text-green-600 font-bold">{d.scheduled_referrals}</span></td>
                      <td className="py-3" onClick={e => e.stopPropagation()}>
                        <button onClick={() => handleDeleteDept(d.id)} className="text-red-500 hover:text-red-700 p-1"><Trash2 className="w-4 h-4" /></button>
                      </td>
                    </tr>

                    {/* EXPANDED ANALYTICS & EDIT FORM */}
                    {expandedDept === d.id && (
                      <tr>
                        <td colSpan="4" className="bg-indigo-50/30 p-5 border-b border-indigo-100">
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

                            {/* Analytics Panel */}
                            <div className="space-y-4">
                              <h4 className="text-xs uppercase font-bold text-indigo-800 tracking-wider">Analytics & Urgences</h4>
                              <div className="flex gap-2 mb-3">
                                <div className="flex-1 bg-white border border-gray-200 p-2 rounded text-center">
                                  <div className="text-[10px] text-gray-500 uppercase font-bold mb-1">Low</div>
                                  <div className="text-lg font-bold text-blue-600">{d.low_urgency}</div>
                                </div>
                                <div className="flex-1 bg-white border border-gray-200 p-2 rounded text-center">
                                  <div className="text-[10px] text-gray-500 uppercase font-bold mb-1">Medium</div>
                                  <div className="text-lg font-bold text-yellow-600">{d.medium_urgency}</div>
                                </div>
                                <div className="flex-1 bg-white border border-gray-200 p-2 rounded text-center">
                                  <div className="text-[10px] text-gray-500 uppercase font-bold mb-1">High</div>
                                  <div className="text-lg font-bold text-orange-600">{d.high_urgency}</div>
                                </div>
                                <div className="flex-1 bg-white border border-gray-200 p-2 rounded text-center">
                                  <div className="text-[10px] text-gray-500 uppercase font-bold mb-1">Critical</div>
                                  <div className="text-lg font-bold text-red-600">{d.critical_urgency}</div>
                                </div>
                              </div>

                              <h4 className="text-xs uppercase font-bold text-indigo-800 tracking-wider flex items-center gap-1 mt-4">
                                <UserCircle2 className="w-4 h-4" /> Docteurs ({d.doctors?.length || 0})
                              </h4>
                              <div className="flex flex-wrap gap-2">
                                {d.doctors?.length > 0 ? d.doctors.map(doc => (
                                  <span key={doc.id} className="text-xs bg-indigo-100 text-indigo-800 px-2 py-1 rounded-full">{doc.username}</span>
                                )) : <span className="text-xs text-gray-400 italic">Aucun docteur assigné</span>}
                              </div>
                            </div>

                            {/* Edit Form */}
                            <div>
                              <h4 className="text-xs uppercase font-bold text-indigo-800 tracking-wider mb-3">Modifier le Service</h4>
                              <form onSubmit={handleUpdateDept} className="space-y-2">
                                <div>
                                  <label className="text-xs text-gray-600 mb-1 block">Nom du Service</label>
                                  <input type="text" value={editDept?.name || ''} onChange={e => setEditDept({ ...editDept, name: e.target.value })} className="w-full border border-gray-300 rounded p-1.5 text-sm" required />
                                </div>
                                <div>
                                  <div>
                                    <label className="text-xs text-gray-600 mb-1 block flex items-center gap-1">
                                      <Phone className="w-3 h-3" /> Numéro De Téléphone
                                    </label>
                                    <input
                                      type="tel"
                                      value={editDept?.phone_extension || ''}
                                      onChange={e => setEditDept({ ...editDept, phone_extension: e.target.value })}
                                      placeholder="+212 6XX XXX XXX"
                                      pattern="[+]?[0-9\s\-\(\)]{10,}"
                                      className="w-full border border-gray-300 rounded p-1.5 text-sm"
                                    />
                                  </div>
                                  <div>
                                    <label className="text-xs text-gray-600 mb-1 block flex items-center gap-1">
                                      <Clock className="w-3 h-3" /> Heures d'Ouverture
                                    </label>
                                    <TimePicker
                                      value={editDept?.work_hours || ''}
                                      onChange={(hours) => setEditDept({ ...editDept, work_hours: hours })}
                                    />
                                  </div>
                                  <div>
                                    <label className="text-xs text-gray-600 mb-1 block">Jours de Travail</label>
                                    <div className="flex gap-1 flex-wrap">
                                      {DAYS.map(day => {
                                        const isSelected = editDept?.work_days?.split(',').includes(day);
                                        return (
                                          <button
                                            type="button"
                                            key={`edit-${day}`}
                                            onClick={() => setEditDept({ ...editDept, work_days: toggleDay(editDept?.work_days, day) })}
                                            className={`px-2.5 py-1 text-xs rounded border transition-colors ${isSelected ? 'bg-indigo-600 border-indigo-600 text-white font-bold shadow-sm' : 'bg-white border-gray-300 text-gray-600 hover:bg-gray-50'}`}
                                          >
                                            {day}
                                          </button>
                                        )
                                      })}
                                    </div>
                                  </div>
                                </div>
                                <button type="submit" className="w-full mt-2 bg-indigo-600 text-white font-bold py-1.5 rounded text-sm hover:bg-indigo-700 flex justify-center items-center gap-1 transition-colors">
                                  <Save className="w-4 h-4" /> Enregistrer
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
          <form onSubmit={handleCreateDept} className="p-5 border-t border-gray-100 bg-gray-50 space-y-3">
            <input type="text" placeholder="Nom du service (ex: Radiologie)" required value={newDept.name} onChange={e => setNewDept({ ...newDept, name: e.target.value })} className="border rounded p-2 text-sm w-full" />

            <div>
              <label className="text-xs text-gray-600 mb-1 block font-medium"><Phone className="w-3 h-3 inline mr-1" />Numéro De Téléphone</label>
              <input type="tel" placeholder="+212 6XX XXX XXX" value={newDept.phone_extension} onChange={e => setNewDept({ ...newDept, phone_extension: e.target.value })} className="border rounded p-2 text-sm w-full" pattern="[+]?[0-9\s\-\(\)]{10,}" />
            </div>

            <div>
              <label className="text-xs text-gray-600 mb-1 block font-medium">Heures d'Ouverture</label>
              <TimePicker
                value={newDept.work_hours || '08:00-16:00'}
                onChange={(hours) => setNewDept({ ...newDept, work_hours: hours })}
              />
            </div>

            <div>
              <label className="text-xs text-gray-600 mb-1 block font-medium">Jours de Travail</label>
              <div className="flex gap-1.5 flex-wrap">
                {DAYS.map(day => {
                  const isSelected = newDept.work_days?.split(',').includes(day);
                  return (
                    <button
                      type="button"
                      key={`new-${day}`}
                      onClick={() => setNewDept({ ...newDept, work_days: toggleDay(newDept.work_days, day) })}
                      className={`px-3 py-1.5 text-[11px] uppercase tracking-wider rounded border transition-colors ${isSelected ? 'bg-indigo-600 border-indigo-600 text-white font-bold shadow-sm' : 'bg-white border-gray-300 text-gray-600 hover:bg-gray-50'}`}
                    >
                      {day}
                    </button>
                  )
                })}
              </div>
            </div>

            <button type="submit" className="w-full bg-indigo-600 text-white font-bold py-2 rounded flex justify-center items-center hover:bg-indigo-700"><Plus className="w-4 h-4 mr-1" /> Ajouter Service</button>
          </form>
        </div>

      </div>
    </div>
  );
}
