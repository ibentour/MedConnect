import { useState, useEffect } from 'react';
import { getDirectory } from '../services/api';
import { Search, MapPin, Phone, Building2, Calendar, Clock } from 'lucide-react';
import { AntiLeak } from '../components/Security';

export default function Directory() {
  const [departments, setDepartments] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDirectory();
  }, []);

  const loadDirectory = async () => {
    try {
      const data = await getDirectory();
      setDepartments(data.departments);
    } catch (error) {
      console.error('Failed to load directory:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredDepts = departments.filter(d => 
    d.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100 mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Annuaire CHU</h1>
        <p className="text-gray-500 mb-6">Rechercher les départements et services disponibles au CHU Mohammed VI</p>
        
        <div className="relative max-w-xl">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-5 w-5 text-gray-400" />
          </div>
          <input
            type="text"
            placeholder="Rechercher par nom de service (ex: Cardiologie)..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="block w-full pl-10 pr-3 py-3 border border-gray-200 rounded-xl leading-5 bg-gray-50 placeholder-gray-400 focus:outline-none focus:bg-white focus:ring-2 focus:ring-brand-500 focus:border-brand-500 transition-colors"
          />
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center p-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-600"></div>
        </div>
      ) : (
        <AntiLeak>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredDepts.map(dept => (
              <div key={dept.id} className="bg-white rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition-shadow overflow-hidden">
                <div className={`h-2 w-full ${dept.is_accepting ? 'bg-green-500' : 'bg-red-500'}`} />
                <div className="p-5">
                  <div className="flex justify-between items-start mb-4">
                    <h3 className="text-lg font-bold text-gray-900 leading-tight">{dept.name}</h3>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      dept.is_accepting ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                    }`}>
                      {dept.is_accepting ? 'Ouvert' : 'Saturé'}
                    </span>
                  </div>
                  
                  <div className="space-y-2.5 text-sm text-gray-600">
                    <div className="flex items-center">
                      <Building2 className="w-4 h-4 mr-2 text-gray-400" />
                      CHU Mohammed VI, Oujda
                    </div>
                    {/* Render exact DB values if present, else fallback logic */}
                    {dept.work_days && (
                      <div className="flex items-center text-blue-700 font-medium">
                        <Calendar className="w-4 h-4 mr-2" />
                        {dept.work_days}
                      </div>
                    )}
                    {dept.work_hours && (
                      <div className="flex items-center text-blue-700 font-medium">
                        <Clock className="w-4 h-4 mr-2" />
                        {dept.work_hours}
                      </div>
                    )}
                    <div className="flex items-center">
                      <Phone className="w-4 h-4 mr-2 text-gray-400" />
                      Ext: {dept.phone_extension || 'N/A'}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
          
          {filteredDepts.length === 0 && (
            <div className="text-center p-12 bg-white rounded-2xl border border-gray-100 border-dashed">
              <p className="text-gray-500">Aucun département trouvé pour "{searchTerm}"</p>
            </div>
          )}
        </AntiLeak>
      )}
    </div>
  );
}
