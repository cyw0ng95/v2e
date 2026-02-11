'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rpcClient } from '@/lib/rpc-client';
import type { ListMemoryCardsRequest, MemoryCard } from '@/lib/types';

export function useMemoryCards(filters: ListMemoryCardsRequest) {
  return useQuery({
    queryKey: ['memory-cards', filters],
    queryFn: () => rpcClient.listMemoryCards(filters || {}),
  });
}

export function useCreateCard() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => rpcClient.createMemoryCard(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory-cards'] });
    },
  });
}

export function useUpdateCard() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ cardId, data }: { cardId: number; data: any }) =>
      rpcClient.updateMemoryCard({ card_id: cardId, ...data }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory-cards'] });
    },
  });
}

export function useDeleteCard() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (cardId: number) => rpcClient.deleteMemoryCard({ card_id: cardId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory-cards'] });
    },
  });
}

export function useRateCard() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { card_id: number; rating: 'again' | 'hard' | 'good' | 'easy' }) =>
      rpcClient.rateMemoryCard(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory-cards'] });
    },
  });
}
