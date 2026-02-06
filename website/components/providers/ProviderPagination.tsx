'use client';

import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface ProviderPaginationProps {
  currentPage: number;
  totalPages: number;
  pageSize: number;
  totalItems?: number;
  onPageChange: (page: number) => void;
}

export function ProviderPagination({
  currentPage,
  totalPages,
  pageSize,
  totalItems,
  onPageChange,
}: ProviderPaginationProps) {
  const startItem = (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(currentPage * pageSize, totalItems || currentPage * pageSize);

  return (
    <div className="flex items-center justify-between border-t pt-4 mt-4">
      <div className="text-sm text-muted-foreground">
        {totalItems !== undefined && (
          <>
            Showing <span className="font-medium">{startItem}</span> to{' '}
            <span className="font-medium">{endItem}</span> of{' '}
            <span className="font-medium">{totalItems}</span> results
          </>
        )}
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Previous
        </Button>
        <div className="text-sm text-muted-foreground px-2">
          Page <span className="font-medium">{currentPage}</span> of{' '}
          <span className="font-medium">{totalPages}</span>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
        >
          Next
          <ChevronRight className="h-4 w-4 ml-1" />
        </Button>
      </div>
    </div>
  );
}
