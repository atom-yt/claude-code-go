import { setupServer } from 'msw/node';
import { handlers } from './handlers';

// MSW server configuration for Node.js environment
export const server = setupServer(...handlers);

// Export handlers for individual test usage if needed
export { handlers };