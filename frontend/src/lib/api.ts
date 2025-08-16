
export const apiClient = {
  request:  async (
    url: string,
    method: 'GET' | 'POST' | 'PATCH' | 'DELETE',
    token: string,
    body?: any
  ) => {
    try {
        const headers: HeadersInit = {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      };

      const options: RequestInit = {
        method, 
        headers, 
        body: body ? JSON.stringify(body) : undefined,
      };

      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}${url}`, options);

      if (!response.ok) {
        let errorData = { message: `HTTP error! status: ${response.status}` };
        try {
          errorData = await response.json();
        } catch (e) {
        }
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
      }

      if (response.status === 204) {
        return null;
      }

      return response.json();

    } catch (error) {
      console.error(`API request failed: ${method} ${url}`, error);
      throw error;
    }
  },

  get: async (url: string, token: string) => apiClient.request(url, 'GET', token),

  post: async (url: string, token: string, body: any) => apiClient.request(url, 'POST', token, body),

  patch: async (url: string, token: string, body: any) => apiClient.request(url, 'PATCH', token, body),
};
