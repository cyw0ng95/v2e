'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { rpcClient } from '@/lib/rpc-client';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { 
  ChevronLeftIcon as ChevronLeft,
  ChevronRightIcon as ChevronRight,
  ListIcon as List,
  LinkIcon as Link
} from '@/components/icons';
import LearningView from './learning-view';

interface LearningNavigationProps {
  initialItemType?: 'CVE' | 'CWE' | 'CAPEC' | 'ATTACK';
  viewMode?: 'view' | 'learn';
}

interface LearningPath {
  currentIndex: number;
  totalItems: number;
  currentURN: string | null;
  strategy: 'BFS' | 'DFS';
  hasNext: boolean;
  hasPrevious: boolean;
}

const LearningNavigation: React.FC<LearningNavigationProps> = ({ 
  initialItemType = 'CWE',
  viewMode = 'learn'
}) => {
  const [items, setItems] = useState<string[]>([]);
  const [currentURN, setCurrentURN] = useState<string | null>(null);
  const [currentIndex, setCurrentIndex] = useState<number>(0);
  const [strategy, setStrategy] = useState<'BFS' | 'DFS'>('BFS');
  const [loading, setLoading] = useState<boolean>(true);
  const [showDetail, setShowDetail] = useState<boolean>(true);

  const loadLearningPath = useCallback(async () => {
    setLoading(true);
    try {
      let urnList: string[] = [];
      
      switch (initialItemType) {
        case 'CVE': {
          const response = await rpcClient.listCVEs(0, 100);
          if (response.retcode === 0 && response.payload) {
            urnList = response.payload.cves.map((cve: any) => 
              `v2e::cve::${cve.id}`
            );
          }
          break;
        }
        case 'CWE': {
          const response = await rpcClient.listCWEs({ limit: 100 });
          if (response.retcode === 0 && response.payload) {
            urnList = response.payload.cwes.map((cwe: any) => 
              `v2e::cwe::${cwe.id}`
            );
          }
          break;
        }
        case 'CAPEC': {
          const response = await rpcClient.listCAPECs(0, 100);
          if (response.retcode === 0 && response.payload) {
            urnList = response.payload.capecs.map((capec: any) => 
              `v2e::capec::${capec.id}`
            );
          }
          break;
        }
        case 'ATTACK': {
          const response = await rpcClient.listAttackTechniques(0, 100);
          if (response.retcode === 0 && response.payload) {
            urnList = response.payload.techniques.map((tech: any) => 
              `v2e::attack::${tech.id}`
            );
          }
          break;
        }
      }
      
      setItems(urnList);
      if (urnList.length > 0) {
        setCurrentURN(urnList[0]);
        setCurrentIndex(0);
      }
    } catch (err) {
      console.error('Error loading learning path:', err);
    } finally {
      setLoading(false);
    }
  }, [initialItemType]);

  useEffect(() => {
    loadLearningPath();
  }, [loadLearningPath]);

  const navigateToNext = useCallback(() => {
    if (currentIndex < items.length - 1) {
      const nextIndex = currentIndex + 1;
      setCurrentIndex(nextIndex);
      setCurrentURN(items[nextIndex]);
    }
  }, [currentIndex, items]);

  const navigateToPrevious = useCallback(() => {
    if (currentIndex > 0) {
      const prevIndex = currentIndex - 1;
      setCurrentIndex(prevIndex);
      setCurrentURN(items[prevIndex]);
    }
  }, [currentIndex, items]);

  const navigateToIndex = useCallback((index: number) => {
    if (index >= 0 && index < items.length) {
      setCurrentIndex(index);
      setCurrentURN(items[index]);
    }
  }, [items]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center text-gray-500">
            No items available for {initialItemType} learning
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Badge variant="outline" className="text-sm">
                {strategy} Strategy
              </Badge>
              <div className="text-sm text-gray-600">
                <span className="font-medium">{currentIndex + 1}</span> of <span className="font-medium">{items.length}</span>
              </div>
            </div>
            
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={navigateToPrevious}
                disabled={currentIndex === 0}
              >
                <ChevronLeft className="w-4 h-4 mr-1" />
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={navigateToNext}
                disabled={currentIndex === items.length - 1}
              >
                Next
                <ChevronRight className="w-4 h-4 ml-1" />
              </Button>
            </div>
          </div>
          
          {viewMode === 'learn' && (
            <div className="mt-4 pt-4 border-t">
              <p className="text-xs text-gray-500">
                Learning strategy is automatically managed. Navigate through items sequentially or follow related links.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {currentURN && (
        <LearningView 
          urn={currentURN}
          onClose={() => setShowDetail(false)}
        />
      )}

      <Card>
        <CardContent className="p-4">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-medium text-gray-700">
              Learning Path
            </h3>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDetail(!showDetail)}
            >
              {showDetail ? <List className="w-4 h-4" /> : <Link className="w-4 h-4" />}
            </Button>
          </div>
          
          <div className="flex space-x-2 overflow-x-auto pb-2">
            {items.map((item, index) => (
              <Button
                key={item}
                variant={index === currentIndex ? "default" : "outline"}
                size="sm"
                onClick={() => navigateToIndex(index)}
                className="flex-shrink-0"
              >
                {index + 1}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default LearningNavigation;
