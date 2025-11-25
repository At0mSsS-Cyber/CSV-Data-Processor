import React from 'react';

function DataTable({ records, isSearchResult = false }) {
  if (!records || records.length === 0) {
    return <div className="text-center py-12 text-gray-500">No records to display</div>;
  }

  // Get all unique fields from all records
  const allFields = new Set();
  records.forEach(record => {
    Object.keys(record.cleanedData).forEach(field => allFields.add(field));
  });
  const fields = Array.from(allFields);

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
            {fields.map(field => (
              <th key={field} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{field}</th>
            ))}
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Grouped Category</th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {records.map(record => (
            <tr key={record.id} className="hover:bg-gray-50">
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{record.id}</td>
              {fields.map(field => (
                <td key={field} className="px-6 py-4 text-sm">
                  <div className="space-y-1">
                    <span className="text-gray-900 font-medium">
                      {record.cleanedData[field] || '-'}
                    </span>
                    {record.originalData[field] !== record.cleanedData[field] && (
                      <span className="block text-xs text-gray-500 italic" title="Original value">
                        ({record.originalData[field]})
                      </span>
                    )}
                  </div>
                </td>
              ))}
              <td className="px-6 py-4 whitespace-nowrap">
                {record.groupedCategory ? (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 border border-gray-200">
                    {record.groupedCategory}
                  </span>
                ) : (
                  <span className="text-gray-400">-</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 text-center text-sm text-gray-600">
        Showing {records.length} record{records.length !== 1 ? 's' : ''}
      </div>
    </div>
  );
}

export default DataTable;
