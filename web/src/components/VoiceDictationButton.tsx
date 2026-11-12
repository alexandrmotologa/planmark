import React, { useState, useEffect, useRef } from 'react';
import { Mic, MicOff } from 'lucide-react';

interface VoiceDictationButtonProps {
  onTranscript: (text: string) => void;
  className?: string;
}

export const VoiceDictationButton: React.FC<VoiceDictationButtonProps> = ({
  onTranscript,
  className = '',
}) => {
  const [isListening, setIsListening] = useState(false);
  const [isSupported, setIsSupported] = useState(false);
  const recognitionRef = useRef<any>(null);

  useEffect(() => {
    const SpeechRecognition =
      (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;

    if (SpeechRecognition) {
      setIsSupported(true);
      const recognition = new SpeechRecognition();
      recognition.continuous = true;
      recognition.interimResults = false;
      recognition.lang = 'en-US';

      recognition.onresult = (event: any) => {
        let text = '';
        for (let i = event.resultIndex; i < event.results.length; ++i) {
          if (event.results[i].isFinal) {
            text += event.results[i][0].transcript + ' ';
          }
        }
        if (text) {
          onTranscript(text.trim());
        }
      };

      recognition.onerror = () => {
        setIsListening(false);
      };

      recognition.onend = () => {
        setIsListening(false);
      };

      recognitionRef.current = recognition;
    }
  }, [onTranscript]);

  const toggleListening = () => {
    if (!recognitionRef.current) return;

    if (isListening) {
      recognitionRef.current.stop();
      setIsListening(false);
    } else {
      try {
        recognitionRef.current.start();
        setIsListening(true);
      } catch {
        setIsListening(false);
      }
    }
  };

  if (!isSupported) return null;

  return (
    <button
      type="button"
      onClick={toggleListening}
      className={`p-1.5 rounded transition flex items-center gap-1 ${
        isListening
          ? 'bg-rose-950/80 text-rose-400 border border-rose-600 animate-pulse'
          : 'bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white border border-gray-700'
      } ${className}`}
      title={isListening ? 'Click to stop voice recording' : 'Dictate feedback using voice (Web Speech API)'}
    >
      {isListening ? (
        <>
          <Mic className="w-3.5 h-3.5 text-rose-400" />
          <span className="text-[10px] font-medium text-rose-300">Listening...</span>
        </>
      ) : (
        <>
          <MicOff className="w-3.5 h-3.5" />
          <span className="text-[10px]">Dictate</span>
        </>
      )}
    </button>
  );
};
