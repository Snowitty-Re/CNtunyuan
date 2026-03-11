import { useRef, useState } from 'react';
import { Button, Slider, Space, Typography } from 'antd';
import { PlayCircleOutlined, PauseCircleOutlined, SoundOutlined } from '@ant-design/icons';
import { formatDuration } from '@/utils/format';

interface Props {
  src: string;
  duration?: number;
}

export default function AudioPlayer({ src, duration: totalDuration }: Props) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);
  const [current, setCurrent] = useState(0);
  const [duration, setDuration] = useState(totalDuration || 0);

  const toggle = () => {
    if (!audioRef.current) return;
    if (playing) {
      audioRef.current.pause();
    } else {
      audioRef.current.play();
    }
    setPlaying(!playing);
  };

  const onTimeUpdate = () => {
    if (audioRef.current) setCurrent(Math.floor(audioRef.current.currentTime));
  };

  const onLoadedMetadata = () => {
    if (audioRef.current) setDuration(Math.floor(audioRef.current.duration));
  };

  const onEnded = () => setPlaying(false);

  const onSeek = (val: number) => {
    if (audioRef.current) {
      audioRef.current.currentTime = val;
      setCurrent(val);
    }
  };

  return (
    <Space align="center" style={{ width: '100%' }}>
      <audio
        ref={audioRef}
        src={src}
        onTimeUpdate={onTimeUpdate}
        onLoadedMetadata={onLoadedMetadata}
        onEnded={onEnded}
      />
      <Button
        type="text"
        icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
        onClick={toggle}
        size="large"
      />
      <SoundOutlined />
      <Slider
        min={0}
        max={duration || 1}
        value={current}
        onChange={onSeek}
        tooltip={{ formatter: (v) => formatDuration(v || 0) }}
        style={{ flex: 1, minWidth: 120 }}
      />
      <Typography.Text type="secondary" style={{ minWidth: 60, textAlign: 'right' }}>
        {formatDuration(current)}/{formatDuration(duration)}
      </Typography.Text>
    </Space>
  );
}
