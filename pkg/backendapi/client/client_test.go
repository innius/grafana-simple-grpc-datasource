package client

//
//type mockClient struct {
//	pb.GrafanaQueryAPIClient
//	mock.Mock
//}
//
//func (m *mockClient) ListMetrics(ctx context.Context, in *pb.ListMetricsRequest, opts ...grpc.CallOption) (*pb.ListMetricsResponse, error) {
//	args := m.Called(ctx, in)
//	return nil, args.Error(1)
//}

// TestListMetricsWithAPIKey tests if secure context handler func is executed
//func TestListMetricsWithAPIKey(t *testing.T) {
//	m := &mockClient{}
//	m.On("ListMetrics", mock.Anything, mock.Anything).Return(nil, nil)
//
//	settings := BackendAPIDatasourceSettings{
//		ID:       "1",
//		Endpoint: "testing-endpoint",
//		APIKey:   "bliep",
//	}
//
//	pool, err := grpcpool.New(clientConnFactory(settings), 1, 1, 10)
//	require.NoError(t, err)
//
//	invocations := 0
//	c := backendAPIClient{
//		pool: pool,
//	}
//	err = c.invoke(context.TODO(), func(ctx context.Context, client pb.GrafanaQueryAPIClient) error {
//		_, err := m.ListMetrics(ctx, &pb.ListMetricsRequest{})
//		return err
//	})
//	_, err = c.ListMetrics(context.TODO(), &pb.ListMetricsRequest{})
//
//	assert.NoError(t, err)
//}

//func TestInvokeBackendWithContext(t *testing.T) {
//	t.Run("without secure context handler", func(t *testing.T) {
//		invocations := 0
//		c := backendAPIClient{}
//		c.invoke(context.TODO(), func(ctx context.Context) {
//			invocations++
//		})
//		assert.Equal(t, 1, invocations)
//	})
//	invocations := 0
//	c := backendAPIClient{
//		secureContext: func(ctx context.Context) context.Context {
//			invocations++
//			return ctx
//		},
//	}
//	c.invoke(context.TODO(), func(ctx context.Context) {
//		invocations++
//	})
//	assert.Equal(t, 2, invocations)
//}
//
//func TestContextWithAPIKey(t *testing.T) {
//	ctx := context.TODO()
//	apiKey := "super-secret-api-key"
//	secureCtx := apiKeySecureContext(apiKey)(ctx)
//
//	md, ok := metadata.FromOutgoingContext(secureCtx)
//	assert.True(t, ok)
//	res := md.Get(apiKeyHeader)
//	if assert.Len(t, res, 1) {
//		assert.Equal(t, apiKey,res[0])
//	}
//}
