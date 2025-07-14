import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import StreamingConfigEditor from '../StreamingConfigEditor';
import { StreamingConfig } from '../../types';

describe('StreamingConfigEditor', () => {
    const defaultConfig: StreamingConfig = {
        maxBufferSize: 3600,
        lookBackPeriod: "1m",
    };

    const mockOnChange = jest.fn();

    beforeEach(() => {
        mockOnChange.mockClear();
    });

    it('renders with default values', () => {
        render(<StreamingConfigEditor config={defaultConfig} onChange={mockOnChange} />);

        expect(screen.getByDisplayValue('3600')).toBeInTheDocument();
        expect(screen.getByDisplayValue('1m')).toBeInTheDocument();
    });

    it('calls onChange when max buffer size changes', () => {
        render(<StreamingConfigEditor config={defaultConfig} onChange={mockOnChange} />);

        const bufferSizeInput = screen.getByDisplayValue('3600');
        fireEvent.change(bufferSizeInput, { target: { value: '5000' } });

        expect(mockOnChange).toHaveBeenCalledWith({
            ...defaultConfig,
            maxBufferSize: 5000,
        });
    });

    it('calls onChange when look-back period changes', () => {
        render(<StreamingConfigEditor config={defaultConfig} onChange={mockOnChange} />);

        const lookBackInput = screen.getByDisplayValue('1m');
        fireEvent.change(lookBackInput, { target: { value: '5m' } });

        expect(mockOnChange).toHaveBeenCalledWith({
            ...defaultConfig,
            lookBackPeriod: '5m',
        });
    });

    it('disables inputs when disabled prop is true', () => {
        render(<StreamingConfigEditor config={defaultConfig} onChange={mockOnChange} disabled={true} />);

        const bufferSizeInput = screen.getByDisplayValue('3600');
        const lookBackInput = screen.getByDisplayValue('1m');

        expect(bufferSizeInput).toBeDisabled();
        expect(lookBackInput).toBeDisabled();
    });

    it('does not call onChange for invalid values', () => {
        render(<StreamingConfigEditor config={defaultConfig} onChange={mockOnChange} />);

        const bufferSizeInput = screen.getByDisplayValue('3600');
        fireEvent.change(bufferSizeInput, { target: { value: 'invalid' } });

        expect(mockOnChange).not.toHaveBeenCalled();
    });
});
