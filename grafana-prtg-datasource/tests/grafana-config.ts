// Grafana e2e test environment setup
export const GRAFANA_CONFIG = {
    url: process.env.GRAFANA_URL || 'http://localhost:3004',
    version: '11.4',
    credentials: {
      username: process.env.GF_SECURITY_ADMIN_USER || 'admin',
      password: process.env.GF_SECURITY_ADMIN_PASSWORD || 'admin'
    },
    apiKey: process.env.GRAFANA_API_KEY || ''
  };
  
  // Helper function to verify Grafana is ready
  export async function verifyGrafanaReady() {
    try {
      const response = await fetch(`${GRAFANA_CONFIG.url}/api/health`);
      if (!response.ok) {
        throw new Error(`Grafana health check failed: ${response.status}`);
      }
      console.log('Grafana is ready');
      return true;
    } catch (error) {
      console.error('Grafana is not ready:', error);    return false;
    }
  }
  