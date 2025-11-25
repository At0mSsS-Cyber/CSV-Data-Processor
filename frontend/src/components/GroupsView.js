import React, { useState, useEffect, useRef } from 'react';

function GroupsView({ groups, fileId }) {
  const [expandedGroups, setExpandedGroups] = useState(new Set());
  const [groupRecords, setGroupRecords] = useState({}); // groupName -> {records: [], page: number, hasMore: bool}
  const [loadingRecords, setLoadingRecords] = useState({});
  const observerTargets = useRef({});
  const RECORDS_PER_PAGE = 20;

  const toggleGroup = async (groupName) => {
    const newExpanded = new Set(expandedGroups);
    if (newExpanded.has(groupName)) {
      newExpanded.delete(groupName);
    } else {
      newExpanded.add(groupName);
      // Load first page of records if not already loaded
      if (!groupRecords[groupName]) {
        await loadGroupRecords(groupName, 1);
      }
    }
    setExpandedGroups(newExpanded);
  };

  const loadGroupRecords = async (groupName, page) => {
    setLoadingRecords(prev => ({ ...prev, [groupName]: true }));

    try {
      const response = await fetch(
        `/api/groups/records?fileId=${fileId}&group=${encodeURIComponent(groupName)}&page=${page}&perPage=${RECORDS_PER_PAGE}`
      );
      const data = await response.json();

      setGroupRecords(prev => ({
        ...prev,
        [groupName]: {
          records: page === 1 ? data.records : [...(prev[groupName]?.records || []), ...data.records],
          page: page,
          hasMore: data.hasMore,
          totalCount: data.totalCount
        }
      }));
    } catch (error) {
      console.error('Error loading group records:', error);
    } finally {
      setLoadingRecords(prev => ({ ...prev, [groupName]: false }));
    }
  };

  const loadMoreRecords = (groupName) => {
    const groupData = groupRecords[groupName];
    if (!groupData || loadingRecords[groupName]) return;
    loadGroupRecords(groupName, groupData.page + 1);
  };

  // Intersection observer for infinite scroll within groups
  useEffect(() => {
    const observers = {};
    
    expandedGroups.forEach(groupName => {
      const target = observerTargets.current[groupName];
      if (!target) return;

      const observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting && groupRecords[groupName]?.hasMore) {
            loadMoreRecords(groupName);
          }
        },
        { threshold: 0.1 }
      );

      observer.observe(target);
      observers[groupName] = observer;
    });

    return () => {
      Object.values(observers).forEach(observer => observer.disconnect());
    };
  }, [expandedGroups, groupRecords, loadingRecords]);

  const groupNames = Object.keys(groups).sort();

  if (groupNames.length === 0) {
    return <div className="text-center py-12 text-gray-500">No grouped categories found</div>;
  }

  return (
    <div className="space-y-4 p-6">
      {groupNames.map(groupName => {
        const recordCount = groups[groupName].length;
        const isExpanded = expandedGroups.has(groupName);
        const groupData = groupRecords[groupName];
        const displayRecords = groupData?.records || [];
        const isLoading = loadingRecords[groupName];

        return (
          <div key={groupName} className="border border-gray-200 rounded-lg overflow-hidden bg-white shadow-sm">
            <div 
              className="px-6 py-4 bg-gradient-to-r from-gray-50 to-white cursor-pointer hover:from-gray-100 hover:to-gray-50 transition-all"
              onClick={() => toggleGroup(groupName)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <span className="text-gray-500">
                    {isExpanded ? (
                      <svg className="w-5 h-5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                      </svg>
                    ) : (
                      <svg className="w-5 h-5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    )}
                  </span>
                  <h4 className="text-lg font-semibold text-gray-900 capitalize">{groupName}</h4>
                </div>
                <span className="inline-flex items-center px-3 py-1.5 rounded-full text-sm font-semibold bg-gray-900 text-white shadow-sm">
                  {recordCount.toLocaleString()} records
                </span>
              </div>
            </div>

            {isExpanded && (
              <div className="bg-white border-t border-gray-200">
                {isLoading && displayRecords.length === 0 ? (
                  <div className="px-6 py-12 text-center">
                    <div className="flex flex-col items-center justify-center space-y-3">
                      <svg className="animate-spin h-8 w-8 text-gray-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <span className="text-gray-600 font-medium">Loading records...</span>
                    </div>
                  </div>
                ) : displayRecords.length === 0 ? (
                  <div className="px-6 py-8 text-center text-gray-500">No records to display</div>
                ) : (
                  <div className="divide-y divide-gray-100">
                    {displayRecords.map(record => (
                      <div key={record.id} className="px-6 py-4 hover:bg-gray-50 transition-colors">
                        <div className="flex gap-4">
                          <div className="flex-shrink-0">
                            <span className="inline-flex items-center justify-center w-10 h-10 rounded-full bg-gray-100 text-gray-700 font-semibold text-sm">
                              #{record.id}
                            </span>
                          </div>
                          <div className="flex-1 space-y-2">
                            {Object.entries(record.cleanedData).map(([field, value]) => (
                              <div key={field} className="flex flex-wrap gap-2 text-sm">
                                <span className="font-semibold text-gray-600 capitalize min-w-[100px]">{field}:</span>
                                <span className="text-gray-900">{value}</span>
                                {record.originalData[field] !== value && (
                                  <span className="text-xs text-gray-400 italic" title="Original value">
                                    (was: {record.originalData[field]})
                                  </span>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    ))}
                    
                    {groupData?.hasMore && (
                      <div 
                        ref={el => observerTargets.current[groupName] = el}
                        className="px-6 py-4 text-center bg-gray-50"
                      >
                        {isLoading ? (
                          <div className="flex items-center justify-center space-x-2 text-gray-600">
                            <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            <span className="text-sm">Loading more...</span>
                          </div>
                        ) : null}
                      </div>
                    )}
                    
                    {displayRecords.length > 0 && !groupData?.hasMore && (
                      <div className="px-6 py-3 text-center text-sm text-gray-500 bg-gray-50">
                        All {displayRecords.length} records loaded
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

export default GroupsView;
