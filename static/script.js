let ntpOffset = 0;
let networkDelay = 0;
let isSync = false;
let currentServer = 'time.cloud.tencent.com';

// 启动时间显示
function startTimeDisplay() {
    setInterval(() => {
        const now = new Date();
        const localTime = formatTime(now);
        
        let ntpTime;
        if (isSync) {
            const ntpNow = new Date(now.getTime() + ntpOffset);
            ntpTime = formatTime(ntpNow);
        } else {
            ntpTime = '--:--:--.---';
        }
        
        document.getElementById('localTime').textContent = localTime;
        document.getElementById('ntpTime').textContent = ntpTime;
    }, 10); // 每10毫秒更新一次
}

function formatTime(date) {
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    const milliseconds = String(date.getMilliseconds()).padStart(3, '0');
    return `${hours}:${minutes}:${seconds}.${milliseconds}`;
}

async function syncTime() {
    const serverUrl = document.getElementById('serverUrl').value.trim();
    const syncBtn = document.getElementById('syncBtn');
    const status = document.getElementById('status');
    const offsetInfo = document.getElementById('offsetInfo');
    const syncInfo = document.getElementById('syncInfo');
    
    // 显示同步状态
    syncBtn.disabled = true;
    syncBtn.innerHTML = '<span class="loading"></span>同步中...';
    status.className = 'status syncing';
    status.textContent = '正在连接NTP服务器...';
    
    try {
        const response = await fetch('/api/sync', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                serverUrl: serverUrl || 'time.cloud.tencent.com'
            })
        });
        
        const result = await response.json();
        
        if (result.success) {
            // 同步成功
            ntpOffset = result.offset;
            networkDelay = result.delay;
            isSync = true;
            
            status.className = 'status success';
            status.textContent = '✓ 同步成功';
            
            // 显示偏移信息
            const offsetClass = result.offset >= 0 ? 'offset-positive' : 'offset-negative';
            const offsetSign = result.offset >= 0 ? '+' : '';
            
            offsetInfo.innerHTML = `
                <div class="offset-value ${offsetClass}">${offsetSign}${result.offset}ms</div>
                <div class="delay-info">延迟: ${result.delay}ms</div>
            `;
            
            // 显示详细信息
            syncInfo.innerHTML = `
                <h3>同步信息</h3>
                <p><strong>服务器:</strong> ${result.serverUrl}</p>
                <p><strong>时间偏移:</strong> ${offsetSign}${result.offset} 毫秒 ${result.offset > 0 ? '(本地时间慢)' : result.offset < 0 ? '(本地时间快)' : '(时间准确)'}</p>
                <p><strong>网络延迟:</strong> ${result.delay} 毫秒</p>
                <p><strong>同步时间:</strong> ${new Date().toLocaleString()}</p>
            `;
            
        } else {
            // 同步失败
            isSync = false;
            status.className = 'status error';
            status.textContent = '✗ ' + result.error;
            
            offsetInfo.innerHTML = '';
            syncInfo.innerHTML = `
                <h3>同步失败</h3>
                <p><strong>错误信息:</strong> ${result.error}</p>
                <p>请检查网络连接或尝试其他NTP服务器</p>
            `;
        }
        
    } catch (error) {
        // 网络错误
        isSync = false;
        status.className = 'status error';
        status.textContent = '✗ 网络连接失败';
        
        offsetInfo.innerHTML = '';
        syncInfo.innerHTML = `
            <h3>连接失败</h3>
            <p><strong>错误信息:</strong> ${error.message}</p>
            <p>请检查网络连接</p>
        `;
    } finally {
        // 恢复按钮状态
        syncBtn.disabled = false;
        syncBtn.textContent = '同步时间';
    }
}

// 选择服务器
function selectServer(serverUrl) {
    currentServer = serverUrl;
    document.getElementById('serverUrl').value = serverUrl;
    
    // 更新按钮状态
    document.querySelectorAll('.server-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    // 自动同步
    setTimeout(syncTime, 300);
}

// 同步系统时间
async function syncSystemTime() {
    const serverUrl = document.getElementById('serverUrl').value.trim();
    const syncSystemBtn = document.getElementById('syncSystemBtn');
    const status = document.getElementById('status');
    
    // 确认对话框
    if (!confirm('WARNING: This will modify system time!\n\nPlease ensure:\n1. Program is running as administrator\n2. Important applications are closed\n3. Confirm to sync to NTP time\n\nContinue?')) {
        return;
    }
    
    // 显示同步状态
    syncSystemBtn.disabled = true;
    syncSystemBtn.innerHTML = '<span class="loading"></span>Syncing system time...';
    status.className = 'status syncing';
    status.textContent = 'Syncing system time, please wait...';
    
    try {
        const response = await fetch('/api/sync-system', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify({
                serverUrl: serverUrl || 'time.cloud.tencent.com'
            })
        });
        
        // 检查响应是否成功
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        // 获取响应文本并尝试解析JSON
        const responseText = await response.text();
        console.log('Response text:', responseText); // 调试信息
        
        let result;
        try {
            result = JSON.parse(responseText);
        } catch (parseError) {
            throw new Error(`Invalid JSON response: ${responseText.substring(0, 100)}`);
        }
        
        if (result.success) {
            status.className = 'status success';
            status.textContent = '✓ ' + (result.message || 'System time synchronized successfully');
            
            // 刷新页面时间显示
            setTimeout(() => {
                location.reload();
            }, 2000);
            
        } else {
            status.className = 'status error';
            status.textContent = '✗ ' + (result.error || 'System time sync failed');
            
        }
        
    } catch (error) {
        console.error('System sync error:', error); // 调试信息
        status.className = 'status error';
        status.textContent = '✗ System time sync failed: ' + error.message;
    } finally {
        syncSystemBtn.disabled = false;
        syncSystemBtn.innerHTML = '同步到系统时间';
    }
}

// 页面加载完成后启动
document.addEventListener('DOMContentLoaded', function() {
    startTimeDisplay();
    
    // 回车键同步
    document.getElementById('serverUrl').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            syncTime();
        }
    });
    
    // 设置默认选中的服务器按钮
    const defaultBtn = document.querySelector(`[onclick="selectServer('time.cloud.tencent.com')"]`);
    if (defaultBtn) {
        defaultBtn.classList.add('active');
    }
    
    // 添加警告提示
    const warningDiv = document.createElement('div');
    warningDiv.className = 'warning-text';
    warningDiv.innerHTML = '系统时间同步需要管理员权限，请右键选择"以管理员身份运行"程序';
    document.querySelector('.action-buttons').appendChild(warningDiv);
    
    // 检测是否有管理员权限
    if (navigator.userAgent.includes('Windows')) {
        warningDiv.classList.add('show');
    }
    
    // 默认同步一次
    setTimeout(syncTime, 1000);

});
