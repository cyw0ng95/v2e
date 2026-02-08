import { describe, it, expect } from 'vitest';
import { rpcClient } from '../src/rpc-client.js';
import { assertRpcSuccess, assertNotFound } from '../helpers/assertions.js';

describe('Memory Cards', () => {
  // Note: Not resetting database between tests as services keep old connections
  // Tests work with the database state from the test setup

  describe('CRUD Operations', () => {
    it('should create a memory card', async () => {
      const response = await rpcClient.call(
        'RPCCreateMemoryCard',
        {
          front_content: 'What is CWE-79?',
          back_content: 'Cross-site Scripting',
          major_class: 'CWE',
          minor_class: 'Injection',
          status: 'active',
          content: '{}',
          card_type: 'basic',
        },
        'local'
      );

      await assertRpcSuccess(response);
      const card = response.payload.memoryCard;
      expect(card.id).toBeDefined();
      expect(card.majorClass).toBe('CWE');
    });

    it('should list memory cards', async () => {
      // First create a card
      await rpcClient.call(
        'RPCCreateMemoryCard',
        {
          front_content: 'Test card',
          back_content: 'Test answer',
          major_class: 'TEST',
          minor_class: 'Test',
          status: 'active',
          content: '{}',
        },
        'local'
      );

      const response = await rpcClient.call(
        'RPCListMemoryCards',
        { limit: 10 },
        'local'
      );

      await assertRpcSuccess(response);
      const data = response.payload as any;
      expect(Array.isArray(data.memoryCards)).toBe(true);
      expect(data.memoryCards.length).toBeGreaterThan(0);
    });

    it('should get a memory card by ID', async () => {
      const createResponse = await rpcClient.call(
        'RPCCreateMemoryCard',
        {
          front_content: 'Get test',
          back_content: 'Answer',
          major_class: 'TEST',
          minor_class: 'Test',
          status: 'active',
          content: '{}',
        },
        'local'
      );

      const cardId = createResponse.payload.memoryCard.id;

      const response = await rpcClient.call(
        'RPCGetMemoryCard',
        { id: cardId },
        'local'
      );

      await assertRpcSuccess(response);
      expect(response.payload.memoryCard.id).toBe(cardId);
    });

    it('should delete a memory card', async () => {
      const createResponse = await rpcClient.call(
        'RPCCreateMemoryCard',
        {
          front_content: 'Delete me',
          back_content: 'Answer',
          major_class: 'TEST',
          minor_class: 'Test',
          status: 'active',
          content: '{}',
        },
        'local'
      );

      const cardId = createResponse.payload.memoryCard.id;

      const deleteResponse = await rpcClient.call(
        'RPCDeleteMemoryCard',
        { id: cardId },
        'local'
      );

      await assertRpcSuccess(deleteResponse);
      expect(deleteResponse.payload.success).toBe(true);

      const getResponse = await rpcClient.call(
        'RPCGetMemoryCard',
        { id: cardId },
        'local'
      );
      assertNotFound(getResponse);
    });
  });

  describe('Rating Operations', () => {
    it('should rate a card with good rating', async () => {
      const createResponse = await rpcClient.call(
        'RPCCreateMemoryCard',
        {
          front_content: 'SM-2 Test',
          back_content: 'Answer',
          major_class: 'TEST',
          minor_class: 'Test',
          status: 'active',
          content: '{}',
        },
        'local'
      );

      const cardId = createResponse.payload.memoryCard.id;

      const rateResponse = await rpcClient.call(
        'RPCRateMemoryCard',
        {
          card_id: cardId,
          rating: 'good',
        },
        'local'
      );

      await assertRpcSuccess(rateResponse);
      const card = rateResponse.payload.memoryCard;
      expect(card.repetition).toBe(1);
      expect(card.fSMState).toBeDefined();
    });

    it('should reset on again rating', async () => {
      const createResponse = await rpcClient.call(
        'RPCCreateMemoryCard',
        {
          front_content: 'Again test',
          back_content: 'Answer',
          major_class: 'TEST',
          minor_class: 'Test',
          status: 'active',
          content: '{}',
        },
        'local'
      );

      const cardId = createResponse.payload.memoryCard.id;

      // First rate as good
      await rpcClient.call(
        'RPCRateMemoryCard',
        { card_id: cardId, rating: 'good' },
        'local'
      );

      // Then rate as again
      const againResponse = await rpcClient.call(
        'RPCRateMemoryCard',
        { card_id: cardId, rating: 'again' },
        'local'
      );

      await assertRpcSuccess(againResponse);
      const card = againResponse.payload.memoryCard;
      expect(card.interval).toBe(1);
      expect(card.fSMState).toBe('to-review');
    });
  });
});
