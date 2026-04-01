import {useEffect, useState} from 'react';
import {GetBatteryStatus, GetConfig, SaveConfig} from '../wailsjs/go/main/App';
import {main} from '../wailsjs/go/models';
import {WindowHide, Quit} from '../wailsjs/runtime/runtime';
import './App.css';

type BatterySnapshot = {
    ok: boolean;
    error?: string;
    deviceName?: string;
    percentage: number;
    state?: string;
    stateLabel?: string;
    onBattery: boolean;
    acOnline: boolean;
    technology?: string;
    capacity?: number;
    energy?: number;
    energyRate?: number;
    voltage?: number;
    chargeCycles?: number;
    warningLevel?: string;
    updatedAgoText?: string;
};

type StatCard = {
    label: string;
    value: string;
    hint: string;
};

const fallbackSnapshot: BatterySnapshot = {
    ok: false,
    percentage: 0,
    onBattery: false,
    acOnline: false,
};

function clampPercentage(value: number) {
    if (Number.isNaN(value)) {
        return 0;
    }
    return Math.max(0, Math.min(100, value));
}

function formatNumber(value?: number, maximumFractionDigits = 1) {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return '—';
    }
    return new Intl.NumberFormat('en-US', {
        maximumFractionDigits,
    }).format(value);
}

function App() {
    const [snapshot, setSnapshot] = useState<BatterySnapshot>(fallbackSnapshot);
    const [isLoading, setIsLoading] = useState(true);
    const [view, setView] = useState<'dashboard' | 'settings'>('dashboard');
    const [config, setConfig] = useState<main.AppConfig>({upperLimit: 80, lowerLimit: 20, intervalMinutes: 10});

    useEffect(() => {
        let active = true;

        const loadConfig = async () => {
            try {
                const conf = await GetConfig();
                if (active && conf) setConfig(conf);
            } catch (err) {
                console.error("Failed to load config", err);
            }
        };

        const loadBattery = async () => {
            try {
                const raw = await GetBatteryStatus();
                const parsed = JSON.parse(raw) as BatterySnapshot;

                if (!active) {
                    return;
                }

                setSnapshot(parsed);
                setIsLoading(false);
            } catch (error) {
                if (!active) {
                    return;
                }

                setSnapshot({
                    ...fallbackSnapshot,
                    ok: false,
                    error: 'Unable to load battery data',
                });
                setIsLoading(false);
                console.error('Failed to load battery snapshot', error);
            }
        };

        loadConfig();
        loadBattery();
        const interval = window.setInterval(loadBattery, 10000);

        return () => {
            active = false;
            window.clearInterval(interval);
        };
    }, []);

    const handleSaveConfig = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await SaveConfig(config);
            setView('dashboard');
        } catch (err) {
            console.error("Failed to save config", err);
        }
    };

    const percentage = clampPercentage(snapshot.percentage);
    const isCharging = snapshot.state === 'charging';
    const isDischarging = snapshot.state === 'discharging';
    
    let ringColor = '#3fb950'; 
    if (percentage <= 20 && !isCharging) ringColor = '#f85149'; 
    else if (percentage <= 50 && !isCharging) ringColor = '#d29922'; 

    const ringStyle = {
        background: `conic-gradient(${ringColor} ${percentage * 3.6}deg, transparent 0deg)`,
    };

    let statusText = 'Unknown';
    if (snapshot.stateLabel) statusText = snapshot.stateLabel;
    else if (snapshot.acOnline && percentage >= 99) statusText = 'Fully Charged';
    else if (snapshot.acOnline) statusText = 'Connected to power';
    else if (snapshot.onBattery) statusText = 'Running on battery';

    const statCards: StatCard[] = [
        {
            label: 'Energy',
            value: `${formatNumber(snapshot.energy)} Wh`,
            hint: 'Current stored charge',
        },
        {
            label: 'Rate',
            value: `${formatNumber(snapshot.energyRate)} W`,
            hint: isCharging ? 'Charging speed' : 'Discharge speed',
        },
        {
            label: 'Voltage',
            value: `${formatNumber(snapshot.voltage, 2)} V`,
            hint: 'Battery voltage',
        },
        {
            label: 'Cycles',
            value: typeof snapshot.chargeCycles === 'number' ? snapshot.chargeCycles.toString() : '—',
            hint: 'Charge cycle count',
        },
        {
            label: 'Capacity',
            value: typeof snapshot.capacity === 'number' ? `${snapshot.capacity.toFixed(0)}%` : '—',
            hint: 'Health indicator',
        },
        {
            label: 'Power State',
            value: snapshot.acOnline ? 'AC Power' : 'Battery',
            hint: snapshot.warningLevel && snapshot.warningLevel !== 'none' 
                ? snapshot.warningLevel.charAt(0).toUpperCase() + snapshot.warningLevel.slice(1) 
                : 'Current source',
        },
    ];

    return (
        <div className="app-container" style={{"--wails-draggable": "drag"} as React.CSSProperties}>
            <header className="titlebar">
                <span className="title">Custos / System Battery</span>
                <div style={{ marginLeft: 'auto', display: 'flex', gap: '8px', zIndex: 10 }}>
                    <button style={{"--wails-draggable": "no-drag", background: 'transparent', border: 'none', color: '#888', cursor: 'pointer', fontSize: '13px'} as React.CSSProperties} onClick={() => setView(view === 'dashboard' ? 'settings' : 'dashboard')} title="Settings">⚙️</button>
                    <button style={{"--wails-draggable": "no-drag", background: 'transparent', border: 'none', color: '#ed5565', cursor: 'pointer', fontSize: '16px', fontWeight: 'bold'} as React.CSSProperties} onClick={WindowHide} title="Hide to Background">×</button>
                </div>
            </header>
            
            <main className="main-content" style={{"--wails-draggable": "no-drag"} as React.CSSProperties}>
                {view === 'settings' ? (
                    <div className="settings-panel">
                        <h2 className="hero-title" style={{marginBottom: '24px'}}>Notification Settings</h2>
                        <form onSubmit={handleSaveConfig} style={{display: 'flex', flexDirection: 'column', gap: '20px'}}>
                            
                            <div className="form-group">
                                <div className="form-group-header">
                                    <label>Upper Limit Threshold (%)</label>
                                </div>
                                <span className="stat-hint" style={{marginTop: '-6px'}}>Alert me when battery charges up to this limit.</span>
                                <div className="form-input-row">
                                    <input type="range" value={config.upperLimit} onChange={e => setConfig({...config, upperLimit: Number(e.target.value)})} min="0" max="100" />
                                    <input type="number" value={config.upperLimit} onChange={e => setConfig({...config, upperLimit: Number(e.target.value)})} min="0" max="100"/>
                                </div>
                            </div>

                            <div className="form-group">
                                <div className="form-group-header">
                                    <label>Lower Limit Threshold (%)</label>
                                </div>
                                <span className="stat-hint" style={{marginTop: '-6px'}}>Alert me when battery discharges down to this limit.</span>
                                <div className="form-input-row">
                                    <input type="range" value={config.lowerLimit} onChange={e => setConfig({...config, lowerLimit: Number(e.target.value)})} min="0" max="100" />
                                    <input type="number" value={config.lowerLimit} onChange={e => setConfig({...config, lowerLimit: Number(e.target.value)})} min="0" max="100"/>
                                </div>
                            </div>

                            <div className="form-group">
                                <div className="form-group-header">
                                    <label>Alert Interval (Minutes)</label>
                                </div>
                                <span className="stat-hint" style={{marginTop: '-6px'}}>Minimum wait time before re-sending the same notification.</span>
                                <div className="form-input-row">
                                    <input type="range" value={config.intervalMinutes} onChange={e => setConfig({...config, intervalMinutes: Number(e.target.value)})} min="1" max="60" />
                                    <input type="number" value={config.intervalMinutes} onChange={e => setConfig({...config, intervalMinutes: Number(e.target.value)})} min="1" max="60"/>
                                </div>
                            </div>

                            <button type="submit" className="save-btn" style={{marginTop: '8px'}}>Save Preferences</button>
                        </form>
                    </div>
                ) : (
                    <>
                        <section className="hero-section">
                            <div className="battery-ring-container">
                                <div className="battery-ring">
                                    <div className="battery-ring-fill" style={ringStyle}></div>
                                    <div className="battery-ring-value">{percentage.toFixed(0)}%</div>
                                </div>
                            </div>

                            <div className="hero-info">
                                <h1 className="hero-title">{snapshot.deviceName || 'Battery Status'}</h1>
                                <p className="hero-subtitle">{snapshot.technology ? `${snapshot.technology} Battery` : 'Internal Power'}</p>
                                
                                <div className={`status-badge ${isCharging ? 'charging' : isDischarging ? 'discharging' : ''}`}>
                                    <span className="status-dot">●</span>
                                    <span>{isLoading ? 'Syncing...' : statusText}</span>
                                </div>
                            </div>
                        </section>

                        <div className="stats-grid">
                            {statCards.map((stat, i) => (
                                <div className="stat-item" key={i}>
                                    <span className="stat-label">{stat.label}</span>
                                    <span className="stat-value">{stat.value}</span>
                                    <span className="stat-hint">{stat.hint}</span>
                                </div>
                            ))}
                        </div>
                    </>
                )}
            </main>
        </div>
    );
}

export default App;
