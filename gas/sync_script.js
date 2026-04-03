/**
 * Google Apps Script — Multi-Sheet Push Sync
 *
 * This script runs inside the Google Spreadsheet.
 * It provides a custom "Sync" menu that sends each sheet's data
 * as a batch JSON POST to the Go backend webhook, with entity_type
 * automatically determined from the sheet tab name.
 *
 * Tab names MUST match the actual spreadsheet tab names:
 * Tasks, Sales, Expenses, Shipments, Sources, Quality, Market Price,
 * Meetings, P&L Forecast, Projects, Partners, Blending
 *
 * Setup:
 * 1. Open the Google Spreadsheet → Extensions → Apps Script
 * 2. Paste this entire script
 * 3. Set the BACKEND_URL and API_SECRET in Script Properties
 * 4. Save & reload the spreadsheet — a "⚡ Sync" menu will appear
 */

// ─── Sheet Name → entity_type mapping (matches Go entity registry keys) ────

var SHEET_ENTITY_MAP = {
    'Tasks': 'tasks',
    'Sales': 'sales',
    'Expenses': 'expenses',
    'Shipments': 'shipments',
    'Sources': 'sources',
    'Quality': 'quality',
    'Market Price': 'market_price',
    'Meetings': 'meetings',
    'P&L Forecast': 'pl_forecast',
    'Projects': 'projects',
    'Partners': 'partners',
    'Blending': 'blending'
};

// ─── Configuration ──────────────────────────────────────────────

function getConfig() {
    var props = PropertiesService.getScriptProperties();
    return {
        backendUrl: (props.getProperty('BACKEND_URL') || 'http://localhost:8080') + '/api/webhook/sheet-sync',
        apiSecret: props.getProperty('API_SECRET') || ''
    };
}

// ─── Menu Setup ─────────────────────────────────────────────────

function onOpen() {
    SpreadsheetApp.getUi()
        .createMenu('⚡ Sync')
        .addItem('Push Current Sheet → DB', 'syncCurrentSheet')
        .addItem('Push ALL Sheets → DB', 'syncAllSheets')
        .addSeparator()
        .addItem('Full Push (DB → All Sheets)', 'triggerFullSync')
        .addItem('Full Pull (All Sheets → DB)', 'triggerPullSync')
        .addToUi();
}

// ─── Sync Current Sheet ────────────────────────────────────────

function syncCurrentSheet() {
    var ui = SpreadsheetApp.getUi();
    var sheet = SpreadsheetApp.getActiveSpreadsheet().getActiveSheet();
    var sheetName = sheet.getName();

    var entityType = SHEET_ENTITY_MAP[sheetName];
    if (!entityType) {
        ui.alert('⚠️ Error', 'Sheet "' + sheetName + '" tidak terdaftar.\n\nSheet yang valid:\n' + Object.keys(SHEET_ENTITY_MAP).join('\n'), ui.ButtonSet.OK);
        return;
    }

    var result = syncSheet(sheet, entityType);
    if (result.success) {
        ui.alert('✅ Sukses', 'Batch ' + result.count + ' item dari "' + sheetName + '" diterima untuk diproses.', ui.ButtonSet.OK);
    } else {
        ui.alert('⚠️ Error', result.message, ui.ButtonSet.OK);
    }
}

// ─── Sync ALL Sheets ───────────────────────────────────────────

function syncAllSheets() {
    var ui = SpreadsheetApp.getUi();
    var spreadsheet = SpreadsheetApp.getActiveSpreadsheet();
    var results = [];

    for (var sheetName in SHEET_ENTITY_MAP) {
        var sheet = spreadsheet.getSheetByName(sheetName);
        if (!sheet) {
            results.push('⏭️ ' + sheetName + ': sheet tidak ditemukan');
            continue;
        }

        var entityType = SHEET_ENTITY_MAP[sheetName];
        var result = syncSheet(sheet, entityType);
        if (result.success) {
            results.push('✅ ' + sheetName + ': ' + result.count + ' item');
        } else {
            results.push('❌ ' + sheetName + ': ' + result.message);
        }
    }

    ui.alert('Hasil Sync', results.join('\n'), ui.ButtonSet.OK);
}

// ─── Core Sync Function ───────────────────────────────────────

function syncSheet(sheet, entityType) {
    var config = getConfig();

    if (!config.apiSecret) {
        return { success: false, message: 'API_SECRET belum diset di Script Properties.' };
    }

    var data = sheet.getDataRange().getValues();

    if (data.length <= 1) {
        return { success: false, message: 'Tidak ada data untuk disinkronkan.' };
    }

    // Row 0 = headers
    var headers = data[0].map(function (h) { return String(h).trim(); });

    var items = [];
    for (var i = 1; i < data.length; i++) {
        var row = data[i];

        // Skip empty rows
        if (!row[0] || String(row[0]).trim() === '') continue;

        var item = {};
        for (var j = 0; j < headers.length; j++) {
            var val = row[j];
            if (val instanceof Date) {
                item[headers[j]] = val.toISOString();
            } else {
                item[headers[j]] = String(val !== null && val !== undefined ? val : '').trim();
            }
        }
        items.push(item);
    }

    if (items.length === 0) {
        return { success: false, message: 'Tidak ada baris valid.' };
    }

    var payload = JSON.stringify({
        entity_type: entityType,
        items: items
    });

    var signature = computeHmacSha256(payload, config.apiSecret);

    var options = {
        method: 'post',
        contentType: 'application/json',
        headers: { 'X-Signature': signature },
        payload: payload,
        muteHttpExceptions: true
    };

    try {
        var response = UrlFetchApp.fetch(config.backendUrl, options);
        var code = response.getResponseCode();

        if (code === 202) {
            return { success: true, count: items.length };
        } else {
            return { success: false, message: 'HTTP ' + code + ': ' + response.getContentText() };
        }
    } catch (e) {
        return { success: false, message: 'Request gagal: ' + e.toString() };
    }
}

// ─── Full Push (DB → All Sheets) ──────────────────────────────

function triggerFullSync() {
    var ui = SpreadsheetApp.getUi();
    var config = getConfig();

    if (!config.apiSecret) {
        ui.alert('Error', 'API_SECRET belum diset.', ui.ButtonSet.OK);
        return;
    }

    var result = ui.alert('Full Push', 'Menulis ulang SEMUA data dari DB ke spreadsheet.\nLanjutkan?', ui.ButtonSet.YES_NO);
    if (result !== ui.Button.YES) return;

    var fullSyncUrl = config.backendUrl.replace('/sheet-sync', '/full-sync');
    sendSyncRequest(fullSyncUrl, config.apiSecret, 'Full push');
}

// ─── Full Pull (All Sheets → DB) ─────────────────────────────

function triggerPullSync() {
    var ui = SpreadsheetApp.getUi();
    var config = getConfig();

    if (!config.apiSecret) {
        ui.alert('Error', 'API_SECRET belum diset.', ui.ButtonSet.OK);
        return;
    }

    var result = ui.alert('Full Pull', 'Sync SEMUA data dari spreadsheet ke database.\nLanjutkan?', ui.ButtonSet.YES_NO);
    if (result !== ui.Button.YES) return;

    var pullSyncUrl = config.backendUrl.replace('/sheet-sync', '/pull-sync');
    sendSyncRequest(pullSyncUrl, config.apiSecret, 'Full pull');
}

// ─── Helper: Send Sync Request ────────────────────────────────

function sendSyncRequest(url, secret, labelPrefix) {
    var ui = SpreadsheetApp.getUi();
    var payload = JSON.stringify({});
    var signature = computeHmacSha256(payload, secret);

    var options = {
        method: 'post',
        contentType: 'application/json',
        headers: { 'X-Signature': signature },
        payload: payload,
        muteHttpExceptions: true
    };

    try {
        var response = UrlFetchApp.fetch(url, options);
        var code = response.getResponseCode();

        if (code === 202) {
            ui.alert('✅ Sukses', labelPrefix + ' telah dijadwalkan.', ui.ButtonSet.OK);
        } else {
            ui.alert('⚠️ Error ' + code, response.getContentText(), ui.ButtonSet.OK);
        }
    } catch (e) {
        ui.alert('❌ Request Gagal', e.toString(), ui.ButtonSet.OK);
    }
}

// ─── HMAC Helper ───────────────────────────────────────────────

function computeHmacSha256(message, secret) {
    var signature = Utilities.computeHmacSha256Signature(message, secret);
    return signature.map(function (byte) {
        return ('0' + (byte & 0xFF).toString(16)).slice(-2);
    }).join('');
}

// ─── Auto Sync on Edit (Installable Trigger) ───────────────────

/**
 * Pushes only the edited row to the database immediately.
 * 
 * IMPORTANT: This requires an INSTALLABLE TRIGGER to work because UrlFetchApp 
 * requires authorization. 
 * Steps to enable:
 * 1. Go to Triggers (clock icon on the left menu)
 * 2. Click "Add Trigger"
 * 3. Choose which function to run: autoSyncOnEdit
 * 4. Select event source: From spreadsheet
 * 5. Select event type: On edit
 * 6. Save and authorize.
 */
function autoSyncOnEdit(e) {
    if (!e || !e.range) return;

    var sheet = e.source.getActiveSheet();
    var sheetName = sheet.getName();

    var entityType = SHEET_ENTITY_MAP[sheetName];
    if (!entityType) return; // Not a managed sheet

    var rowIdx = e.range.getRow();
    if (rowIdx <= 1) return; // Header row or invalid

    var dataRange = sheet.getDataRange();
    // Re-verify we have headers
    var firstRow = dataRange.getValues()[0];
    if (!firstRow) return;

    var headers = firstRow.map(function (h) { return String(h).trim(); });

    // Get only the edited row
    var rowValues = sheet.getRange(rowIdx, 1, 1, headers.length).getValues()[0];

    // Skip empty rows or rows without an ID
    if (!rowValues[0] || String(rowValues[0]).trim() === '') return;

    var item = {};
    for (var j = 0; j < headers.length; j++) {
        var val = rowValues[j];
        if (val instanceof Date) {
            item[headers[j]] = val.toISOString();
        } else {
            item[headers[j]] = String(val !== null && val !== undefined ? val : '').trim();
        }
    }

    var config = getConfig();
    if (!config.apiSecret) return;

    var payload = JSON.stringify({
        entity_type: entityType,
        items: [item] // Batch of 1 item
    });

    var signature = computeHmacSha256(payload, config.apiSecret);

    var options = {
        method: 'post',
        contentType: 'application/json',
        headers: { 'X-Signature': signature },
        payload: payload,
        muteHttpExceptions: true
    };

    try {
        UrlFetchApp.fetch(config.backendUrl, options);
    } catch (err) {
        console.error("Auto-sync failed on row " + rowIdx + ": " + err.toString());
    }
}
