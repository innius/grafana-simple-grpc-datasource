import { DataSource } from '../datasource';
import { DataSourceInstanceSettings } from '@grafana/data';
import { MyDataSourceOptions, MyQuery, QueryType } from '../types';

describe('DataSource - createStreamingPath', () => {
  const settings = {
    jsonData: {
      apikey_authentication_enabled: false,
    },
  } as DataSourceInstanceSettings<MyDataSourceOptions>;

  const ds = new DataSource(settings);

  // Access private method for testing
  const createStreamingPath = (query: MyQuery): string => {
    return (ds as any).createStreamingPath(query);
  };

  const createHash = (input: string): string => {
    return (ds as any).createHash(input);
  };

  describe('hash function', () => {
    it('should produce consistent hashes for same input', () => {
      const input = 'test-string';
      const hash1 = createHash(input);
      const hash2 = createHash(input);
      expect(hash1).toBe(hash2);
    });

    it('should produce different hashes for different inputs', () => {
      const hash1 = createHash('input1');
      const hash2 = createHash('input2');
      expect(hash1).not.toBe(hash2);
    });

    it('should produce compact hash strings', () => {
      const hash = createHash('very-long-input-string-that-would-exceed-limits');
      expect(hash.length).toBeLessThan(20); // Base36 should be compact
    });
  });

  describe('basic path creation', () => {
    it('should handle simple refId only', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
      };
      const path = createStreamingPath(query);
      // QueryType is a meaningful component, so it should be hashed
      expect(path).toContain('A');
      expect(path.length).toBeLessThanOrEqual(100);
      expect(path).toMatch(/^A\/[a-z0-9]+$/); // Should be A/hash format
    });

    it('should handle refId with query components', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricHistory,
        metrics: [{ metricId: 'temperature' }],
      };
      const path = createStreamingPath(query);
      expect(path).toContain('A');
      expect(path.length).toBeLessThanOrEqual(100);
    });

    it('should handle truly minimal query (no queryType)', () => {
      // Create a query without queryType to test the minimal case
      const query = {
        refId: 'A',
      } as MyQuery;
      const path = createStreamingPath(query);
      expect(path).toBe('A');
      expect(path.length).toBeLessThanOrEqual(100);
    });
  });

  describe('length constraint enforcement', () => {
    it('should handle very long refId by truncating', () => {
      const longRefId = 'A'.repeat(150); // Exceeds 100 chars
      const query: MyQuery = {
        refId: longRefId,
        queryType: QueryType.GetMetricValue,
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
    });

    it('should handle complex query that would exceed 100 chars', () => {
      const query: MyQuery = {
        refId: 'VeryLongRefIdThatCouldCauseProblems',
        queryType: QueryType.GetMetricAggregate,
        metrics: [
          { metricId: 'very-long-metric-name-that-adds-significant-length' },
        ],
        dimensions: [
          { key: 'very-long-dimension-key', value: 'very-long-dimension-value', id: '1' },
          { key: 'another-long-key', value: 'another-long-value', id: '2' },
          { key: 'third-dimension', value: 'third-value', id: '3' },
        ],
        queryOptions: {
          'long-option-name': { label: 'Long Option', value: 'very-long-option-value' },
          'another-option': { label: 'Another', value: 'another-long-value' },
        },
        streamingConfig: {
          maxBufferSize: 10000,
          lookBackPeriod: '24h',
        },
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
    });
  });

  describe('path uniqueness', () => {
    it('should generate different paths for different queries', () => {
      const query1: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: 'temperature' }],
      };
      const query2: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: 'humidity' }],
      };
      const path1 = createStreamingPath(query1);
      const path2 = createStreamingPath(query2);
      expect(path1).not.toBe(path2);
    });

    it('should generate same path for identical queries', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricHistory,
        metrics: [{ metricId: 'temperature' }],
        dimensions: [{ key: 'zone', value: 'north', id: '1' }],
      };
      const path1 = createStreamingPath(query);
      const path2 = createStreamingPath(query);
      expect(path1).toBe(path2);
    });

    it('should differentiate based on streaming config', () => {
      const baseQuery: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: 'temperature' }],
      };
      
      const query1 = {
        ...baseQuery,
        streamingConfig: { maxBufferSize: 1000 },
      };
      
      const query2 = {
        ...baseQuery,
        streamingConfig: { maxBufferSize: 2000 },
      };

      const path1 = createStreamingPath(query1);
      const path2 = createStreamingPath(query2);
      expect(path1).not.toBe(path2);
    });
  });

  describe('component handling', () => {
    it('should handle multiple metrics', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricHistory,
        metrics: [
          { metricId: 'temperature' },
          { metricId: 'humidity' },
        ],
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
      // Should be deterministic
      expect(createStreamingPath(query)).toBe(path);
    });

    it('should handle multiple dimensions', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: 'temperature' }],
        dimensions: [
          { key: 'zone', value: 'north', id: '1' },
          { key: 'building', value: 'A', id: '2' },
          { key: 'floor', value: '1', id: '3' },
        ],
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
    });

    it('should handle query options', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricAggregate,
        metrics: [{ metricId: 'temperature' }],
        queryOptions: {
          'aggregation': { label: 'Aggregation', value: 'avg' },
          'interval': { label: 'Interval', value: '5m' },
        },
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
    });

    it('should ignore query options without values', () => {
      const query1: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: 'temperature' }],
        queryOptions: {
          'option1': { label: 'Option 1', value: 'value1' },
          'option2': { label: 'Option 2', value: '' }, // Empty value
        },
      };

      const query2: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: 'temperature' }],
        queryOptions: {
          'option1': { label: 'Option 1', value: 'value1' },
          // option2 not present
        },
      };

      const path1 = createStreamingPath(query1);
      const path2 = createStreamingPath(query2);
      expect(path1).toBe(path2);
    });
  });

  describe('special characters handling', () => {
    it('should sanitize special characters in refId', () => {
      const query: MyQuery = {
        refId: 'A-B@C#D$E%F',
        queryType: QueryType.GetMetricValue,
      };
      const path = createStreamingPath(query);
      expect(path).toMatch(/^[a-zA-Z0-9_\/]+$/); // Only allowed characters
    });

    it('should handle empty or undefined components gracefully', () => {
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: undefined,
        dimensions: undefined,
        queryOptions: undefined,
        streamingConfig: undefined,
      };
      const path = createStreamingPath(query);
      // Even with undefined components, queryType is still present, so it will be hashed
      expect(path).toContain('A');
      expect(path.length).toBeLessThanOrEqual(100);
      expect(path).toMatch(/^A\/[a-z0-9]+$/); // Should be A/hash format
    });

    it('should handle truly empty query components', () => {
      const query = {
        refId: 'A',
        // No queryType or other components
      } as MyQuery;
      const path = createStreamingPath(query);
      expect(path).toBe('A');
      expect(path.length).toBeLessThanOrEqual(100);
    });
  });

  describe('edge cases', () => {
    it('should handle extremely long individual components', () => {
      const longString = 'x'.repeat(200);
      const query: MyQuery = {
        refId: 'A',
        queryType: QueryType.GetMetricValue,
        metrics: [{ metricId: longString }],
        dimensions: [{ key: longString, value: longString, id: '1' }],
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
    });

    it('should handle query with all possible components', () => {
      const query: MyQuery = {
        refId: 'ComplexQuery',
        queryType: QueryType.GetMetricAggregate,
        metrics: [
          { metricId: 'metric1' },
          { metricId: 'metric2' },
        ],
        dimensions: [
          { key: 'dim1', value: 'val1', id: '1' },
          { key: 'dim2', value: 'val2', id: '2' },
        ],
        queryOptions: {
          'opt1': { label: 'Option 1', value: 'value1' },
          'opt2': { label: 'Option 2', value: 'value2' },
        },
        streamingConfig: {
          maxBufferSize: 5000,
          lookBackPeriod: '2h',
        },
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
      
      // Should be consistent
      expect(createStreamingPath(query)).toBe(path);
    });

    it('should handle empty refId', () => {
      const query: MyQuery = {
        refId: '',
        queryType: QueryType.GetMetricValue,
      };
      const path = createStreamingPath(query);
      expect(path.length).toBeLessThanOrEqual(100);
      expect(path.length).toBeGreaterThan(0);
    });
  });

  describe('performance characteristics', () => {
    it('should generate paths quickly for complex queries', () => {
      const complexQuery: MyQuery = {
        refId: 'PerformanceTest',
        queryType: QueryType.GetMetricAggregate,
        metrics: Array.from({ length: 10 }, (_, i) => ({ metricId: `metric${i}` })),
        dimensions: Array.from({ length: 20 }, (_, i) => ({ 
          key: `dimension${i}`, 
          value: `value${i}`, 
          id: i.toString() 
        })),
        queryOptions: Object.fromEntries(
          Array.from({ length: 15 }, (_, i) => [
            `option${i}`, 
            { label: `Option ${i}`, value: `value${i}` }
          ])
        ),
        streamingConfig: {
          maxBufferSize: 10000,
          lookBackPeriod: '24h',
        },
      };

      const startTime = performance.now();
      const path = createStreamingPath(complexQuery);
      const endTime = performance.now();

      expect(path.length).toBeLessThanOrEqual(100);
      expect(endTime - startTime).toBeLessThan(10); // Should be very fast (< 10ms)
    });
  });
});
