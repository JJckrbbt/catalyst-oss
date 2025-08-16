import { useState, useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { apiClient } from '@/lib/api';

interface StatusHistoryEntry {
  status_history_id: number;
  status: string;
  status_date: string;
  notes: string;
  user_id: number;
  user_first_name: string;
  user_last_name: string;
  user_email: string;
}

interface StatusHistoryProps {
  id: number;
  type: 'items';
}

export function StatusHistory({ id, type }: StatusHistoryProps) {
  const [history, setHistory] = useState<StatusHistoryEntry[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { getAccessTokenSilently, isAuthenticated } = useAuth0();

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        setIsLoading(true);
        const token = await getAccessTokenSilently({
          authorizationParams: {
            audience: import.meta.env.VITE_AUTH0_AUDIENCE,
          },
        });
        const url = `/api/items/history/${id}`;
        const data = await apiClient.get(url, token);
        setHistory(data);
      } catch (error) {
        console.error('StatusHistory: Failed to fetch status history:', error);
      } finally {
        setIsLoading(false);
      }
    };

    if (id && isAuthenticated) {
      fetchHistory();
    }
  }, [id, type, isAuthenticated, getAccessTokenSilently]);

  if (isLoading) {
    return <p>Loading status history...</p>;
  }

  if (!history || history.length === 0) {
    return <p>No status history available.</p>;
  }

  return (
    <div className="space-y-4">
      {history.map((entry) => (
        <div key={entry.status_history_id} className="p-2 border rounded-md">
          <p><strong>Status:</strong> {entry.status}</p>
          <p><strong>Date:</strong> {new Date(entry.status_date).toLocaleString()}</p>
          <p><strong>User:</strong> {entry.user_first_name} {entry.user_last_name} ({entry.user_email})</p>
          <p><strong>Notes:</strong> {entry.notes}</p>
        </div>
      ))}
    </div>
  );
}
