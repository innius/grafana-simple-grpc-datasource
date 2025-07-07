import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import StreamingToggle from '../StreamingToggle';

describe('StreamingToggle', () => {
  const mockOnChange = jest.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
  });

  describe('Switch Mode', () => {
    it('renders switch with boolean value', () => {
      render(<StreamingToggle value={true} onChange={mockOnChange} />);
      
      const switchElement = screen.getByRole('switch');
      expect(switchElement).toBeInTheDocument();
      expect(switchElement).toBeChecked();
    });

    it('calls onChange with boolean when switch is toggled', () => {
      render(<StreamingToggle value={false} onChange={mockOnChange} />);
      
      const switchElement = screen.getByRole('switch');
      fireEvent.click(switchElement);
      
      expect(mockOnChange).toHaveBeenCalledWith(true);
    });

    it('shows "Use Variable" button in switch mode', () => {
      render(<StreamingToggle value={true} onChange={mockOnChange} />);
      
      expect(screen.getByText('Use Variable')).toBeInTheDocument();
    });
  });

  describe('Variable Mode', () => {
    it('renders text input when value is a string', () => {
      render(<StreamingToggle value="$streaming" onChange={mockOnChange} />);
      
      const input = screen.getByDisplayValue('$streaming');
      expect(input).toBeInTheDocument();
    });

    it('calls onChange with string when text input changes', () => {
      render(<StreamingToggle value="$streaming" onChange={mockOnChange} />);
      
      const input = screen.getByDisplayValue('$streaming');
      fireEvent.change(input, { target: { value: '$myVariable' } });
      
      expect(mockOnChange).toHaveBeenCalledWith('$myVariable');
    });

    it('shows "Switch to Toggle" button in variable mode', () => {
      render(<StreamingToggle value="$streaming" onChange={mockOnChange} />);
      
      expect(screen.getByText('Switch to Toggle')).toBeInTheDocument();
    });

    it('handles static string values', () => {
      render(<StreamingToggle value="true" onChange={mockOnChange} />);
      
      const input = screen.getByDisplayValue('true');
      expect(input).toBeInTheDocument();
    });
  });

  describe('Mode Switching', () => {
    it('switches from switch to variable mode', () => {
      render(<StreamingToggle value={true} onChange={mockOnChange} />);
      
      const button = screen.getByText('Use Variable');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith('$streaming');
    });

    it('switches from variable to switch mode', () => {
      render(<StreamingToggle value="$streaming" onChange={mockOnChange} />);
      
      const button = screen.getByText('Switch to Toggle');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith(true);
    });

    it('preserves boolean value when switching to variable mode', () => {
      render(<StreamingToggle value={false} onChange={mockOnChange} />);
      
      const button = screen.getByText('Use Variable');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith('$streaming');
    });

    it('evaluates string value when switching to switch mode', () => {
      render(<StreamingToggle value="false" onChange={mockOnChange} />);
      
      const button = screen.getByText('Switch to Toggle');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith(false);
    });
  });

  describe('Value Evaluation', () => {
    it('treats variable references as enabled in switch display', () => {
      render(<StreamingToggle value="$streaming" onChange={mockOnChange} />);
      
      // Switch to toggle mode to see the evaluated state
      const button = screen.getByText('Switch to Toggle');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith(true);
    });

    it('evaluates "true" string as enabled', () => {
      render(<StreamingToggle value="true" onChange={mockOnChange} />);
      
      const button = screen.getByText('Switch to Toggle');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith(true);
    });

    it('evaluates "false" string as disabled', () => {
      render(<StreamingToggle value="false" onChange={mockOnChange} />);
      
      const button = screen.getByText('Switch to Toggle');
      fireEvent.click(button);
      
      expect(mockOnChange).toHaveBeenCalledWith(false);
    });
  });

  describe('Disabled State', () => {
    it('does not call onChange when switch is clicked and disabled', () => {
      render(<StreamingToggle value={true} onChange={mockOnChange} disabled={true} />);
      
      const switchElement = screen.getByRole('switch');
      fireEvent.click(switchElement);
      
      expect(mockOnChange).not.toHaveBeenCalled();
    });

    it('does not call onChange when text input changes and disabled', () => {
      render(<StreamingToggle value="$streaming" onChange={mockOnChange} disabled={true} />);
      
      const input = screen.getByDisplayValue('$streaming');
      fireEvent.change(input, { target: { value: '$newVariable' } });
      
      expect(mockOnChange).not.toHaveBeenCalled();
    });

    it('does not call onChange when mode switch button is clicked and disabled', () => {
      render(<StreamingToggle value={true} onChange={mockOnChange} disabled={true} />);
      
      const button = screen.getByText('Use Variable');
      fireEvent.click(button);
      
      expect(mockOnChange).not.toHaveBeenCalled();
    });
  });
});
