import React, { useState, useEffect, useRef, useCallback } from 'react';
import FileUpload from './components/FileUpload';
import FilesList from './components/FilesList';
import SearchBar from './components/SearchBar';
import DataTable from './components/DataTable';
import GroupsView from './components/GroupsView';

function App() {
  const [selectedFile, setSelectedFile] = useState(null);
  const [data, setData] = useState(null);
  const [allRecords, setAllRecords] = useState([]);
  const [searchResults, setSearchResults] = useState(null);
  const [allSearchResults, setAllSearchResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [view, setView] = useState('all'); // 'all' or 'groups'
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalRecords, setTotalRecords] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [searchPage, setSearchPage] = useState(1);
  const [searchHasMore, setSearchHasMore] = useState(false);
  const RECORDS_PER_PAGE = 100;
  const observerTarget = useRef(null);
  const [groupsData, setGroupsData] = useState(null);
  const [loadingGroups, setLoadingGroups] = useState(false);

  const handleUploadSuccess = (response) => {
    console.log('Upload successful:', response);
    // Trigger refresh of files list
    setRefreshTrigger(prev => prev + 1);
    setSelectedFile(null);
    setData(null);
  };

  const handleSelectFile = async (file) => {
    setSelectedFile(file);
    setSearchResults(null);
    setAllRecords([]);
    setAllSearchResults([]);
    setCurrentPage(1);
    setSearchPage(1);
    setHasMore(false);
    setSearchHasMore(false);
    setGroupsData(null);
    fetchFileData(file.id, 1);
  };

  const fetchFileData = async (fileId, page) => {
    setLoading(true);
    try {
      const response = await fetch(`/api/records?fileId=${fileId}&page=${page}&perPage=${RECORDS_PER_PAGE}`);
      const result = await response.json();
      setData(result);
      
      if (page === 1) {
        setAllRecords(result.records || []);
      } else {
        setAllRecords(prev => [...prev, ...(result.records || [])]);
      }
      
      setTotalRecords(result.totalCount);
      setHasMore(result.hasMore);
      setCurrentPage(page);
    } catch (error) {
      console.error('Error fetching data:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchGroupsData = async (fileId) => {
    if (groupsData) return; // Already loaded
    
    setLoadingGroups(true);
    try {
      const response = await fetch(`/api/records?fileId=${fileId}&page=1&perPage=10000`);
      const result = await response.json();
      setGroupsData(result.groups || {});
    } catch (error) {
      console.error('Error fetching groups:', error);
    } finally {
      setLoadingGroups(false);
    }
  };

  const loadMoreRecords = useCallback(() => {
    if (loadingMore || !hasMore || !selectedFile) return;
    setLoadingMore(true);
    
    fetchFileData(selectedFile.id, currentPage + 1).finally(() => {
      setLoadingMore(false);
    });
  }, [loadingMore, hasMore, selectedFile, currentPage]);

  const handleSearch = async (query) => {
    if (!selectedFile) return;

    setSearchPage(1);
    setAllSearchResults([]);
    performSearch(query, 1);
  };

  const performSearch = async (query, page) => {
    setLoading(true);
    try {
      // Use the unified /api/records endpoint with optional ?q= parameter
      const response = await fetch(`/api/records?fileId=${selectedFile.id}&q=${encodeURIComponent(query)}&page=${page}&perPage=${RECORDS_PER_PAGE}`);
      const result = await response.json();
      
      if (page === 1) {
        setAllSearchResults(result.records || []);
      } else {
        setAllSearchResults(prev => [...prev, ...(result.records || [])]);
      }
      
      setSearchResults({
        query: query,
        count: result.totalCount,
        totalCount: result.totalCount
      });
      setSearchHasMore(result.hasMore);
      setSearchPage(page);
    } catch (error) {
      console.error('Error searching:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadMoreSearchResults = useCallback(() => {
    if (loadingMore || !searchHasMore || !searchResults) return;
    setLoadingMore(true);
    
    performSearch(searchResults.query, searchPage + 1).finally(() => {
      setLoadingMore(false);
    });
  }, [loadingMore, searchHasMore, searchResults, searchPage]);

  const handleClearSearch = () => {
    setSearchResults(null);
    setAllSearchResults([]);
    setSearchPage(1);
    setSearchHasMore(false);
  };

  const handleBackToList = () => {
    setSelectedFile(null);
    setData(null);
    setSearchResults(null);
    setAllRecords([]);
    setAllSearchResults([]);
    setCurrentPage(1);
    setSearchPage(1);
    setHasMore(false);
    setSearchHasMore(false);
  };

  // Infinite scroll observer
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          if (searchResults) {
            loadMoreSearchResults();
          } else if (view === 'all') {
            loadMoreRecords();
          }
        }
      },
      { threshold: 0.1 }
    );

    const currentTarget = observerTarget.current;
    if (currentTarget) {
      observer.observe(currentTarget);
    }

    return () => {
      if (currentTarget) {
        observer.unobserve(currentTarget);
      }
    };
  }, [searchResults, view, loadMoreRecords, loadMoreSearchResults]);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <h1 className="text-3xl font-semibold text-gray-900">CSV Data Processor</h1>
          <p className="mt-1 text-sm text-gray-600">Upload, Clean, Group, and Search Large CSV Files</p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {!selectedFile ? (
          <>
            <div className="mb-8">
              <FileUpload onUploadSuccess={handleUploadSuccess} />
            </div>

            <div>
              <FilesList 
                key={refreshTrigger} 
                onSelectFile={handleSelectFile} 
              />
            </div>
          </>
        ) : (
          <>
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
              <button 
                className="mb-4 inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500" 
                onClick={handleBackToList}
              >
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
                Back to Files
              </button>
              <h2 className="text-xl font-semibold text-gray-900">{selectedFile.filename}</h2>
              <div className="mt-2 flex items-center text-sm text-gray-600 space-x-4">
                <span>Records: {selectedFile.recordCount}</span>
                <span>•</span>
                <span>Processing Time: {selectedFile.processingTimeMs}ms</span>
              </div>
            </div>

            <div className="mb-6">
              <SearchBar 
                onSearch={handleSearch} 
                onClear={handleClearSearch}
              />
            </div>

            {data && (
              <>
                <div className="flex space-x-2 mb-6">
                  <button 
                    className={`px-4 py-2 text-sm font-medium rounded-lg border ${
                      view === 'all' 
                        ? 'bg-gray-900 text-white border-gray-900' 
                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                    }`}
                    onClick={() => setView('all')}
                  >
                    All Records
                  </button>
                  <button 
                    className={`px-4 py-2 text-sm font-medium rounded-lg border ${
                      view === 'groups' 
                        ? 'bg-gray-900 text-white border-gray-900' 
                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                    }`}
                    onClick={() => {
                      setView('groups');
                      if (selectedFile) {
                        fetchGroupsData(selectedFile.id);
                      }
                    }}
                  >
                    Grouped View
                  </button>
                </div>

                <div className="bg-white rounded-lg shadow-sm border border-gray-200">
                  {loading && !loadingMore ? (
                    <div className="text-center py-12 text-gray-600">Loading...</div>
                  ) : searchResults ? (
                    <>
                      <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
                        <h3 className="text-lg font-semibold text-gray-900">
                          Search Results ({searchResults.totalCount} found) - Showing {allSearchResults.length}
                        </h3>
                        <button 
                          onClick={handleClearSearch} 
                          className="px-3 py-1.5 text-sm font-medium text-red-700 bg-red-50 border border-red-200 rounded-md hover:bg-red-100"
                        >
                          Clear Search
                        </button>
                      </div>
                      <DataTable records={allSearchResults} isSearchResult={true} />
                      {searchHasMore && (
                        <div ref={observerTarget} className="px-6 py-4 border-t border-gray-200 text-center">
                          {loadingMore && (
                            <div className="flex items-center justify-center space-x-2 text-gray-600">
                              <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                              </svg>
                              <span>Loading more results...</span>
                            </div>
                          )}
                        </div>
                      )}
                    </>
                  ) : view === 'all' ? (
                    <>
                      <div className="px-6 py-4 border-b border-gray-200">
                        <h3 className="text-lg font-semibold text-gray-900">
                          All Records ({totalRecords}) - Showing {allRecords.length}
                        </h3>
                      </div>
                      <DataTable records={allRecords} />
                      {hasMore && (
                        <div ref={observerTarget} className="px-6 py-4 border-t border-gray-200 text-center">
                          {loadingMore && (
                            <div className="flex items-center justify-center space-x-2 text-gray-600">
                              <svg className="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                              </svg>
                              <span>Loading more records... ({totalRecords - allRecords.length} remaining)</span>
                            </div>
                          )}
                        </div>
                      )}
                    </>
                  ) : (
                    <>
                      <div className="px-6 py-4 border-b border-gray-200">
                        <h3 className="text-lg font-semibold text-gray-900">Grouped Categories</h3>
                      </div>
                      {loadingGroups ? (
                        <div className="text-center py-12 text-gray-600">
                          <div className="flex items-center justify-center space-x-2">
                            <svg className="animate-spin h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            <span>Loading groups...</span>
                          </div>
                        </div>
                      ) : groupsData ? (
                        <GroupsView groups={groupsData} fileId={selectedFile.id} />
                      ) : (
                        <div className="text-center py-12 text-gray-500">No groups data available</div>
                      )}
                    </>
                  )}
                </div>
              </>
            )}
          </>
        )}
      </main>

      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-center text-sm text-gray-500">Built with Go + PostgreSQL + React | ElSapien Assessment</p>
        </div>
      </footer>
    </div>
  );
}

export default App;
