import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { FeedService } from '../gen/api/v1/pods_connect'

// Import service definition that you want to connect to.
// import { ElizaService } from '@buf/connectrpc_eliza.connectrpc_es/connectrpc/eliza/v1/eliza_connect';
// import { FeedService } from '../gen/api/v1/pods_pb';

// The transport defines what type of endpoint we're hitting.
// In our example we'll be communicating with a Connect endpoint.
// If your endpoint only supports gRPC-web, make sure to use
// `createGrpcWebTransport` instead.
const transport = createConnectTransport({
	baseUrl: 'http://localhost:8080',
	// useBinaryFormat: /[?&]json=0/.test(globalThis.location?.search || '')
	useBinaryFormat: false,
})

// Here we make the client itself, combining the service
// definition with the transport.
const apiClient = createPromiseClient(FeedService, transport)

export default apiClient
