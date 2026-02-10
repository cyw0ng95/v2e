import { describe, it, expect, beforeEach } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertNotFound } from '../helpers/assertions.js';

describe('GLC Canvas', () => {
  // Note: Tests work with the database state from the test setup

  describe('Graph CRUD Operations', () => {
    let createdGraphId: string | null = null;

    it('should create a graph', async () => {
      const response = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Test Graph',
          description: 'A test graph for E2E testing',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{"x":0,"y":0,"zoom":1}',
        },
        'local'
      );

      await assertRpcSuccess(response);
      const graph = response.payload as any;
      expect(graph.graphId).toBeDefined();
      expect(graph.name).toBe('Test Graph');
      expect(graph.presetId).toBe('d3fend');
      createdGraphId = graph.graphId;
    });

    it('should get a graph by ID', async () => {
      // First create a graph
      const createResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Get Test Graph',
          description: 'Graph for get test',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = createResponse.payload.graphId;

      const response = await rpcClient.call(
        'RPCGetGLCGraph',
        { graph_id: graphId },
        'local'
      );

      await assertRpcSuccess(response);
      expect(response.payload.graphId).toBe(graphId);
      expect(response.payload.name).toBe('Get Test Graph');
    });

    it('should update a graph', async () => {
      // Create a graph first
      const createResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Original Name',
          description: '',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = createResponse.payload.graphId;

      // Update the graph
      const response = await rpcClient.call(
        'RPCUpdateGLCGraph',
        {
          graph_id: graphId,
          name: 'Updated Name',
          description: 'Updated description',
        },
        'local'
      );

      await assertRpcSuccess(response);
      expect(response.payload.name).toBe('Updated Name');
      expect(response.payload.description).toBe('Updated description');
    });

    it('should delete a graph', async () => {
      // Create a graph first
      const createResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'To Be Deleted',
          description: '',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = createResponse.payload.graphId;

      // Delete the graph
      const deleteResponse = await rpcClient.call(
        'RPCDeleteGLCGraph',
        { graph_id: graphId },
        'local'
      );

      await assertRpcSuccess(deleteResponse);

      // Verify deletion
      const getResponse = await rpcClient.call(
        'RPCGetGLCGraph',
        { graph_id: graphId },
        'local'
      );

      assertNotFound(getResponse);
    });

    it('should list graphs', async () => {
      // Create a few graphs
      for (let i = 0; i < 3; i++) {
        await rpcClient.call(
          'RPCCreateGLCGraph',
          {
            name: `List Test Graph ${i}`,
            description: '',
            preset_id: 'd3fend',
            nodes: '[]',
            edges: '[]',
            viewport: '{}',
          },
          'local'
        );
      }

      const response = await rpcClient.call(
        'RPCListGLCGraphs',
        { preset_id: 'd3fend', offset: 0, limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);
      const data = response.payload as any;
      expect(Array.isArray(data.graphs)).toBe(true);
      expect(data.graphs.length).toBeGreaterThanOrEqual(3);
    });
  });

  describe('Version Operations', () => {
    it('should create version on graph update with nodes/edges', async () => {
      // Create a graph
      const createResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Version Test Graph',
          description: '',
          preset_id: 'd3fend',
          nodes: '[{"id":"node-1"}]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = createResponse.payload.graphId;

      // Update with new nodes
      await rpcClient.call(
        'RPCUpdateGLCGraph',
        {
          graph_id: graphId,
          nodes: '[{"id":"node-1"},{"id":"node-2"}]',
        },
        'local'
      );

      // Get versions
      const versionsResponse = await rpcClient.call(
        'RPCListGLCGraphVersions',
        { graph_id: graphId, limit: 10 },
        'local'
      );

      await assertRpcSuccess(versionsResponse);
      const data = versionsResponse.payload as any;
      expect(Array.isArray(data.versions)).toBe(true);
      expect(data.versions.length).toBeGreaterThanOrEqual(1);
    });

    it('should restore a graph to a previous version', async () => {
      // Create a graph
      const createResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Restore Test Graph',
          description: '',
          preset_id: 'd3fend',
          nodes: '[{"id":"original-node"}]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = createResponse.payload.graphId;

      // Update with new nodes to create version
      await rpcClient.call(
        'RPCUpdateGLCGraph',
        {
          graph_id: graphId,
          nodes: '[{"id":"new-node"}]',
        },
        'local'
      );

      // Restore to version 1
      const restoreResponse = await rpcClient.call(
        'RPCRestoreGLCGraphVersion',
        { graph_id: graphId, version: 1 },
        'local'
      );

      await assertRpcSuccess(restoreResponse);
      expect(restoreResponse.payload.nodes).toContain('original-node');
    });
  });

  describe('User Preset Operations', () => {
    it('should create a user preset', async () => {
      const response = await rpcClient.call(
        'RPCCreateGLCUserPreset',
        {
          name: 'Test Preset',
          version: '1.0.0',
          description: 'A test preset',
          author: 'E2E Test',
          theme: '{"primary":"#6366f1"}',
          behavior: '{"snapToGrid":true}',
          node_types: '[]',
          relationships: '[]',
        },
        'local'
      );

      await assertRpcSuccess(response);
      expect(response.payload.presetId).toBeDefined();
      expect(response.payload.name).toBe('Test Preset');
    });

    it('should list user presets', async () => {
      // Create a preset first
      await rpcClient.call(
        'RPCCreateGLCUserPreset',
        {
          name: 'List Test Preset',
          version: '1.0.0',
          description: '',
          author: '',
          theme: '{}',
          behavior: '{}',
          node_types: '[]',
          relationships: '[]',
        },
        'local'
      );

      const response = await rpcClient.call(
        'RPCListGLCUserPresets',
        {},
        'local'
      );

      await assertRpcSuccess(response);
      const data = response.payload as any;
      expect(Array.isArray(data.presets)).toBe(true);
      expect(data.presets.length).toBeGreaterThanOrEqual(1);
    });

    it('should delete a user preset', async () => {
      // Create a preset
      const createResponse = await rpcClient.call(
        'RPCCreateGLCUserPreset',
        {
          name: 'To Delete Preset',
          version: '1.0.0',
          description: '',
          author: '',
          theme: '{}',
          behavior: '{}',
          node_types: '[]',
          relationships: '[]',
        },
        'local'
      );

      const presetId = createResponse.payload.presetId;

      // Delete the preset
      const deleteResponse = await rpcClient.call(
        'RPCDeleteGLCUserPreset',
        { preset_id: presetId },
        'local'
      );

      await assertRpcSuccess(deleteResponse);

      // Verify deletion
      const getResponse = await rpcClient.call(
        'RPCGetGLCUserPreset',
        { preset_id: presetId },
        'local'
      );

      assertNotFound(getResponse);
    });
  });

  describe('Share Link Operations', () => {
    it('should create a share link', async () => {
      // Create a graph first
      const graphResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Share Test Graph',
          description: '',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = graphResponse.payload.graphId;

      const response = await rpcClient.call(
        'RPCCreateGLCShareLink',
        {
          graph_id: graphId,
          password: '',
          expires_in_seconds: null,
        },
        'local'
      );

      await assertRpcSuccess(response);
      expect(response.payload.linkId).toBeDefined();
      expect(response.payload.graphId).toBe(graphId);
    });

    it('should get a graph by share link', async () => {
      // Create a graph
      const graphResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Share Access Test Graph',
          description: 'Test share access',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = graphResponse.payload.graphId;

      // Create share link
      const linkResponse = await rpcClient.call(
        'RPCCreateGLCShareLink',
        {
          graph_id: graphId,
          password: '',
          expires_in_seconds: null,
        },
        'local'
      );

      const linkId = linkResponse.payload.linkId;

      // Get graph by share link
      const response = await rpcClient.call(
        'RPCGetGLCGraphByShareLink',
        { link_id: linkId, password: '' },
        'local'
      );

      await assertRpcSuccess(response);
      expect(response.payload.graphId).toBe(graphId);
    });

    it('should require password for password-protected links', async () => {
      // Create a graph
      const graphResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Protected Graph',
          description: '',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = graphResponse.payload.graphId;

      // Create password-protected share link
      const linkResponse = await rpcClient.call(
        'RPCCreateGLCShareLink',
        {
          graph_id: graphId,
          password: 'secret123',
          expires_in_seconds: null,
        },
        'local'
      );

      const linkId = linkResponse.payload.linkId;

      // Try without password
      const wrongResponse = await rpcClient.call(
        'RPCGetGLCGraphByShareLink',
        { link_id: linkId, password: '' },
        'local'
      );

      expect(wrongResponse.retcode).not.toBe(0);

      // Try with correct password
      const correctResponse = await rpcClient.call(
        'RPCGetGLCGraphByShareLink',
        { link_id: linkId, password: 'secret123' },
        'local'
      );

      await assertRpcSuccess(correctResponse);
    });

    it('should delete a share link', async () => {
      // Create a graph
      const graphResponse = await rpcClient.call(
        'RPCCreateGLCGraph',
        {
          name: 'Delete Link Test Graph',
          description: '',
          preset_id: 'd3fend',
          nodes: '[]',
          edges: '[]',
          viewport: '{}',
        },
        'local'
      );

      const graphId = graphResponse.payload.graphId;

      // Create share link
      const linkResponse = await rpcClient.call(
        'RPCCreateGLCShareLink',
        {
          graph_id: graphId,
          password: '',
          expires_in_seconds: null,
        },
        'local'
      );

      const linkId = linkResponse.payload.linkId;

      // Delete the link
      const deleteResponse = await rpcClient.call(
        'RPCDeleteGLCShareLink',
        { link_id: linkId },
        'local'
      );

      await assertRpcSuccess(deleteResponse);

      // Verify deletion
      const getResponse = await rpcClient.call(
        'RPCGetGLCShareLink',
        { link_id: linkId },
        'local'
      );

      assertNotFound(getResponse);
    });
  });
});
