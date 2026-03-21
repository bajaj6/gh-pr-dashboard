const DATA = JSON.parse(document.getElementById('dashboard-data').textContent);

// ============================================================
// Theme management
// ============================================================
var ThemeManager = {
  current: 'system',
  init: function() {
    var saved = 'system';
    try { saved = localStorage.getItem('gh-dashboard-theme') || 'system'; } catch(e) {}
    this.apply(saved);
    this.bindToggle();
  },
  apply: function(choice) {
    this.current = choice;
    try { localStorage.setItem('gh-dashboard-theme', choice); } catch(e) {}
    var effective;
    if (choice === 'system') {
      effective = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } else {
      effective = choice;
    }
    document.documentElement.setAttribute('data-theme', effective);
    var buttons = document.querySelectorAll('.theme-toggle button');
    for (var i = 0; i < buttons.length; i++) {
      var btn = buttons[i];
      if (btn.getAttribute('data-theme-choice') === choice) {
        btn.classList.add('active');
      } else {
        btn.classList.remove('active');
      }
    }
  },
  bindToggle: function() {
    var self = this;
    var toggle = document.getElementById('themeToggle');
    if (toggle) {
      toggle.addEventListener('click', function(e) {
        var btn = e.target.closest('button');
        if (btn && btn.getAttribute('data-theme-choice')) {
          self.apply(btn.getAttribute('data-theme-choice'));
        }
      });
    }
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function() {
      if (self.current === 'system') self.apply('system');
    });
  }
};

// ============================================================
// Collapsed state
// ============================================================
var CollapsedState = {
  collapsed: {},
  init: function() {
    try { this.collapsed = JSON.parse(localStorage.getItem('gh-dashboard-collapsed') || '{}'); } catch(e) { this.collapsed = {}; }
  },
  isCollapsed: function(key) {
    return !!this.collapsed[key];
  },
  toggle: function(key) {
    this.collapsed[key] = !this.collapsed[key];
    try { localStorage.setItem('gh-dashboard-collapsed', JSON.stringify(this.collapsed)); } catch(e) {}
  }
};

// ============================================================
// Bot detection
// ============================================================
var BOT_PATTERNS = [
  /\[bot\]$/i, /^dependabot/i, /^renovate/i, /^github-actions/i,
  /^snyk-/i, /^greenkeeper/i, /^imgbot/i, /^codecov/i,
  /^circleci/i, /^mergify/i, /^linear-app/i, /^allstar-app/i,
  /^depfu/i, /-bot$/i, /^bot-/i
];

var BOT_USERNAMES = DATA.extraBotUsernames || [];

function isBot(pr) {
  var author = getAuthor(pr).toLowerCase();
  for (var i = 0; i < BOT_USERNAMES.length; i++) {
    if (author === BOT_USERNAMES[i].toLowerCase()) return true;
  }
  for (var i = 0; i < BOT_PATTERNS.length; i++) {
    if (BOT_PATTERNS[i].test(author)) return true;
  }
  return false;
}

// ============================================================
// Helpers
// ============================================================
function exclude(source) {
  var excludeSets = Array.prototype.slice.call(arguments, 1);
  var urls = {};
  excludeSets.forEach(function(s) { s.forEach(function(pr) { urls[pr.url] = true; }); });
  return source.filter(function(pr) { return !urls[pr.url]; });
}

function getRepoName(pr) {
  if (pr.repository && pr.repository.nameWithOwner) return pr.repository.nameWithOwner;
  if (pr.repository && pr.repository.name) return pr.repository.name;
  return '';
}
function getAuthor(pr) { return (pr.author && pr.author.login) ? pr.author.login : 'unknown'; }
function getAvatarUrl(pr) { return 'https://github.com/' + getAuthor(pr) + '.png?size=64'; }

function timeAgo(dateStr) {
  var seconds = Math.floor((new Date() - new Date(dateStr)) / 1000);
  if (seconds < 60) return 'now';
  var minutes = Math.floor(seconds / 60);
  if (minutes < 60) return minutes + 'm';
  var hours = Math.floor(minutes / 60);
  if (hours < 24) return hours + 'h';
  var days = Math.floor(hours / 24);
  if (days < 30) return days + 'd';
  var months = Math.floor(days / 30);
  return months + 'mo';
}

function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function labelColor(label) { return label.color ? '#' + label.color : '#6c757d'; }
function textColorForBg(hex) {
  hex = hex.replace('#', '');
  var r = parseInt(hex.substr(0, 2), 16), g = parseInt(hex.substr(2, 2), 16), b = parseInt(hex.substr(4, 2), 16);
  return (r * 299 + g * 587 + b * 114) / 1000 > 140 ? '#1F2328' : '#ffffff';
}

// ============================================================
// Team filter state
// ============================================================
var STORAGE_KEY = 'pr-dashboard-team-filters';
function loadTeamFilters() { try { var s = localStorage.getItem(STORAGE_KEY); return s ? JSON.parse(s) : {}; } catch(e) { return {}; } }
function saveTeamFilters(f) { try { localStorage.setItem(STORAGE_KEY, JSON.stringify(f)); } catch(e) {} }
var teamFilters = loadTeamFilters();
function isTeamEnabled(team) { return teamFilters[team] !== false; }
function toggleTeamFilter(team) { teamFilters[team] = !isTeamEnabled(team); saveTeamFilters(teamFilters); render(); }
function filterByTeams(prs) {
  return prs.filter(function(pr) {
    if (!pr._teams || pr._teams.length === 0) return true;
    return pr._teams.some(function(t) { return isTeamEnabled(t); });
  });
}

// ============================================================
// Time filter
// ============================================================
var TIME_STORAGE_KEY = 'pr-dashboard-time-filter';
var DEFAULT_TIME = '3w';
function loadTimeFilter() { try { var s = localStorage.getItem(TIME_STORAGE_KEY); return s || DEFAULT_TIME; } catch(e) { return DEFAULT_TIME; } }
function saveTimeFilter(v) { try { localStorage.setItem(TIME_STORAGE_KEY, v); } catch(e) {} }
var activeTimeFilter = loadTimeFilter();
// Migrate old numeric format to new string format
if (activeTimeFilter === '1' || activeTimeFilter === 1) activeTimeFilter = '1w';
if (activeTimeFilter === '3' || activeTimeFilter === 3) activeTimeFilter = '3w';
if (activeTimeFilter === '4' || activeTimeFilter === 4) activeTimeFilter = '1m';
if (activeTimeFilter === '0' || activeTimeFilter === 0 || activeTimeFilter === 'all') activeTimeFilter = '3w';

function setTimeFilter(val) { activeTimeFilter = val; saveTimeFilter(val); render(); }
function filterByTime(prs) {
  var weeks;
  if (activeTimeFilter === '1w') weeks = 1;
  else if (activeTimeFilter === '3w') weeks = 3;
  else if (activeTimeFilter === '1m') weeks = 4;
  else return prs;
  var cutoff = new Date(); cutoff.setDate(cutoff.getDate() - (weeks * 7));
  return prs.filter(function(pr) { return new Date(pr.createdAt) >= cutoff; });
}

// ============================================================
// Process data
// ============================================================
function processData() {
  var myApproved = DATA.myApproved;
  var myDraft = DATA.myDraft;
  var myChangesReq = DATA.myChangesReq;
  var myOpen = exclude(DATA.myOpenRaw, myApproved, myDraft, myChangesReq);
  var reviewTeam = DATA.reviewTeam;
  var reviewChangesReq = DATA.reviewChangesReq;
  var approvedTeam = DATA.approvedTeam;
  var reviewDirect = exclude(DATA.reviewDirect, reviewTeam);
  var approvedMe = exclude(DATA.approvedMe, approvedTeam);

  var teamUrlSet = {};
  reviewTeam.forEach(function(pr) { teamUrlSet[pr.url] = true; });
  reviewDirect = reviewDirect.filter(function(pr) { return !teamUrlSet[pr.url]; });

  function notMyPr(pr) { return getAuthor(pr).toLowerCase() !== DATA.user.toLowerCase(); }
  reviewDirect = reviewDirect.filter(notMyPr);
  reviewTeam = reviewTeam.filter(notMyPr);
  reviewChangesReq = reviewChangesReq.filter(notMyPr);
  approvedMe = approvedMe.filter(notMyPr);
  approvedTeam = approvedTeam.filter(notMyPr);

  var allReviewerPrs = [].concat(reviewDirect, reviewTeam, reviewChangesReq, approvedMe, approvedTeam);
  var botPrs = [], draftPrs = [], botUrls = {}, draftUrls = {};
  allReviewerPrs.forEach(function(pr) {
    if (isBot(pr) && !botUrls[pr.url]) { botPrs.push(pr); botUrls[pr.url] = true; }
    if (pr.isDraft && !draftUrls[pr.url] && !botUrls[pr.url]) { draftPrs.push(pr); draftUrls[pr.url] = true; }
  });
  function cleanReviewer(prs) { return prs.filter(function(pr) { return !botUrls[pr.url] && !draftUrls[pr.url]; }); }

  return {
    myApproved: myApproved, myOpen: myOpen, myDraft: myDraft, myChangesReq: myChangesReq,
    reviewDirect: cleanReviewer(reviewDirect), reviewTeam: cleanReviewer(reviewTeam),
    reviewChangesReq: cleanReviewer(reviewChangesReq),
    approvedMe: cleanReviewer(approvedMe), approvedTeam: cleanReviewer(approvedTeam),
    botPrs: botPrs, draftPrs: draftPrs
  };
}

// ============================================================
// Stale detection
// ============================================================
var STALE_HOURS = DATA.staleHours || 24;
function staleAnchor(pr) {
  return pr.reviewRequestedAt || pr.createdAt;
}
function isStale(pr) {
  return (new Date() - new Date(staleAnchor(pr))) > STALE_HOURS * 60 * 60 * 1000;
}

// ============================================================
// SVG Icons
// ============================================================
var ICONS = {
  prOpen: '<svg viewBox="0 0 16 16" width="16" height="16"><path fill="var(--color-pr-open)" d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25Zm5.677-.177L9.573.677A.25.25 0 0 1 10 .854V2.5h1A2.5 2.5 0 0 1 13.5 5v5.628a2.251 2.251 0 1 1-1.5 0V5a1 1 0 0 0-1-1h-1v1.646a.25.25 0 0 1-.427.177L7.177 3.427a.25.25 0 0 1 0-.354ZM3.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm0 9.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm8.25.75a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0Z"></path></svg>',
  prDraft: '<svg viewBox="0 0 16 16" width="16" height="16"><path fill="var(--color-pr-draft)" d="M3.25 1A2.25 2.25 0 0 1 4 5.372v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 3.25 1Zm9.5 5.5a.75.75 0 0 1 .75.75v3.378a2.251 2.251 0 1 1-1.5 0V7.25a.75.75 0 0 1 .75-.75Zm-1.07-4.677a.25.25 0 0 1 0 .354l-.854.853a.75.75 0 0 1-1.06-1.06l.853-.854a.25.25 0 0 1 .354 0l.707.707ZM9.573 5.677a.25.25 0 0 1-.427-.177V3.854a.75.75 0 0 1 1.5 0V5.5a.25.25 0 0 1-.427.177l-.646-.646ZM3.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm0 9.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm8.25.75a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0Z"></path></svg>',
  chevron: '<svg class="subsection-toggle" viewBox="0 0 16 16"><path fill-rule="evenodd" d="M12.78 5.22a.749.749 0 0 1 0 1.06l-4.25 4.25a.749.749 0 0 1-1.06 0L3.22 6.28a.749.749 0 1 1 1.06-1.06L8 8.939l3.72-3.719a.749.749 0 0 1 1.06 0Z"></path></svg>'
};

// ============================================================
// Render functions
// ============================================================
function renderPrIcon(pr) {
  if (pr.isDraft) return ICONS.prDraft;
  return ICONS.prOpen;
}

function renderLabel(label) {
  var bg = labelColor(label);
  var fg = textColorForBg(bg);
  return '<span class="Label" style="background:' + bg + ';color:' + fg + '">' + escapeHtml(label.name) + '</span>';
}

function renderBadges(pr, highlightStale) {
  var html = '';
  if (highlightStale && isStale(pr)) html += '<span class="Badge Badge-stale">Waiting ' + timeAgo(staleAnchor(pr)) + '</span>';
  if (pr.isDraft) html += '<span class="Badge Badge-draft">Draft</span>';
  if (isBot(pr)) html += '<span class="Badge Badge-bot">Bot</span>';
  return html;
}

function renderTeamTags(pr) {
  if (!pr._teams || pr._teams.length === 0) return '';
  return pr._teams.map(function(t) { return '<span class="team-tag">' + escapeHtml(t) + '</span>'; }).join('');
}

function renderPrRow(pr, highlightStale) {
  var repo = getRepoName(pr);
  var repoShort = repo.split('/').pop();
  var author = getAuthor(pr);
  var labelsHtml = '';
  if (pr.labels && pr.labels.length > 0) {
    labelsHtml = pr.labels.map(renderLabel).join('');
  }
  return '<div class="pr-row">' +
    '<div class="pr-status-icon">' + renderPrIcon(pr) + '</div>' +
    '<div class="pr-main">' +
      '<div class="pr-title-row">' +
        '<a class="pr-title" href="' + pr.url + '" target="_blank" rel="noopener noreferrer">' + escapeHtml(pr.title) + '</a>' +
        labelsHtml +
        renderBadges(pr, highlightStale) +
      '</div>' +
      '<div class="pr-meta">' +
        '<span class="pr-repo">' + escapeHtml(repoShort) + '#' + pr.number + '</span>' +
        '<span class="dot">&middot;</span>' +
        '<a class="pr-author" href="https://github.com/' + encodeURIComponent(author) + '" target="_blank" rel="noopener noreferrer"><img class="pr-author-avatar" src="https://github.com/' + encodeURIComponent(author) + '.png?size=40" alt="" />' + escapeHtml(author) + '</a>' +
        '<span class="dot">&middot;</span>' +
        '<span>' + timeAgo(pr.updatedAt) + '</span>' +
        renderTeamTags(pr) +
      '</div>' +
    '</div></div>';
}

function renderSubsection(id, label, prs, countClass, highlightStale) {
  var filtered = filterByTime(prs);
  var isCollapsed = CollapsedState.isCollapsed(id);
  var count = filtered.length;

  var content;
  if (count === 0) {
    content = '<div class="subsection-empty">No pull requests</div>';
  } else {
    var sorted = filtered.slice().sort(function(a, b) { return new Date(a.createdAt) - new Date(b.createdAt); });
    content = sorted.map(function(pr) { return renderPrRow(pr, highlightStale); }).join('');
  }

  return '<div class="subsection ' + (isCollapsed ? 'collapsed' : '') + '" data-section-id="' + id + '">' +
    '<div class="subsection-header" onclick="toggleSection(\'' + id + '\')">' +
      ICONS.chevron +
      '<span class="subsection-label">' + escapeHtml(label) + '</span>' +
      '<span class="subsection-count ' + countClass + '">' + count + '</span>' +
    '</div>' +
    '<div class="subsection-content">' + content + '</div>' +
  '</div>';
}

function renderSectionGroup(title, icon, subsections) {
  var totalCount = 0;
  subsections.forEach(function(s) { totalCount += filterByTime(s.prs).length; });

  return '<div class="section-group">' +
    '<div class="section-group-header">' +
      icon +
      '<span class="section-group-title">' + escapeHtml(title) + '</span>' +
      '<span class="section-group-count">' + totalCount + '</span>' +
    '</div>' +
    '<div class="section-group-content">' +
      subsections.map(function(s) { return renderSubsection(s.id, s.label, s.prs, s.countClass, s.highlightStale); }).join('') +
    '</div>' +
  '</div>';
}

function computeStats(d) {
  var reviewDirect = filterByTime(d.reviewDirect);
  var reviewTeam = filterByTime(filterByTeams(d.reviewTeam));
  var needsReview = reviewDirect.length + reviewTeam.length;

  var approvedMe = filterByTime(d.approvedMe);
  var approvedTeam = filterByTime(filterByTeams(d.approvedTeam));
  var totalApproved = approvedMe.length + approvedTeam.length;

  var myApproved = filterByTime(d.myApproved);
  var myChangesReq = filterByTime(d.myChangesReq);
  var myOpen = filterByTime(d.myOpen);
  var myDraft = filterByTime(d.myDraft);
  var totalMy = myApproved.length + myChangesReq.length + myOpen.length + myDraft.length;

  var allReview = [].concat(reviewDirect, reviewTeam);
  var staleCount = allReview.filter(function(pr) { return isStale(pr); }).length;

  return { needsReview: needsReview, totalApproved: totalApproved, totalMy: totalMy, staleCount: staleCount };
}

function renderStats(stats) {
  var staleHtml = stats.staleCount > 0
    ? ' <span style="color:var(--color-attention-fg);font-size:12px">(Waiting ' + stats.staleCount + ')</span>'
    : '';

  return '<div class="summary-stats">' +
    '<div class="stat-item">' +
      '<span class="stat-icon"><svg viewBox="0 0 16 16"><path fill="var(--color-danger-fg)" d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25Zm5.677-.177L9.573.677A.25.25 0 0 1 10 .854V2.5h1A2.5 2.5 0 0 1 13.5 5v5.628a2.251 2.251 0 1 1-1.5 0V5a1 1 0 0 0-1-1h-1v1.646a.25.25 0 0 1-.427.177L7.177 3.427a.25.25 0 0 1 0-.354Z"></path></svg></span>' +
      'Needs review <span class="stat-counter danger">' + stats.needsReview + '</span>' + staleHtml +
    '</div>' +
    '<div class="stat-item">' +
      '<span class="stat-icon"><svg viewBox="0 0 16 16"><path fill="var(--color-success-fg)" d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"></path></svg></span>' +
      'Approved <span class="stat-counter success">' + stats.totalApproved + '</span>' +
    '</div>' +
    '<div class="stat-item">' +
      '<span class="stat-icon"><svg viewBox="0 0 16 16"><path fill="var(--color-done-fg)" d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25Z"></path></svg></span>' +
      'My PRs <span class="stat-counter done">' + stats.totalMy + '</span>' +
    '</div>' +
  '</div>';
}

function getTeamCounts(d) {
  var counts = {};
  DATA.teams.forEach(function(t) { counts[t] = 0; });
  var allPrs = filterByTime([].concat(d.reviewDirect, d.reviewTeam, d.reviewChangesReq, d.approvedMe, d.approvedTeam, d.botPrs, d.draftPrs, d.myApproved, d.myChangesReq, d.myOpen, d.myDraft));
  allPrs.forEach(function(pr) {
    if (pr._teams) {
      pr._teams.forEach(function(t) {
        if (counts[t] !== undefined) counts[t]++;
      });
    }
  });
  return counts;
}

function renderFilterBar(d) {
  var teamCounts = getTeamCounts(d);

  var timeOptions = [
    { val: '1w', label: '1 week' },
    { val: '3w', label: '3 weeks' },
    { val: '1m', label: '1 month' }
  ];
  var timeButtons = timeOptions.map(function(b) {
    return '<button class="' + (activeTimeFilter === b.val ? 'selected' : '') + '" onclick="setTimeFilter(\'' + b.val + '\')">' + b.label + '</button>';
  }).join('');

  var teamPills = '';
  if (DATA.teams.length > 0) {
    var sortedTeams = DATA.teams.slice()
      .filter(function(t) { return (teamCounts[t] || 0) > 0; })
      .sort(function(a, b) { return (teamCounts[b] || 0) - (teamCounts[a] || 0); });
    teamPills = sortedTeams.map(function(t) {
      var active = isTeamEnabled(t);
      var shortName = t.split('/')[1] || t;
      return '<button class="team-pill ' + (active ? 'active' : '') + '" onclick="toggleTeamFilter(\'' + escapeHtml(t) + '\')">' +
        escapeHtml(shortName) +
        '<span class="team-count">' + (teamCounts[t] || 0) + '</span>' +
      '</button>';
    }).join('');
  }

  var html = '<div class="filter-bar">' +
    '<div class="filter-group"><div class="time-nav">' + timeButtons + '</div></div>';

  if (teamPills) {
    html += '<div class="filter-separator"></div>' +
      '<div class="filter-group">' +
        '<span class="filter-group-label">Teams</span>' +
        '<div class="team-filters">' + teamPills + '</div>' +
      '</div>';
  }

  return html + '</div>';
}

// ============================================================
// Main render
// ============================================================
function render() {
  var d = processData();
  var reviewTeamFiltered = filterByTeams(d.reviewTeam);
  var approvedTeamFiltered = filterByTeams(d.approvedTeam);
  var botFiltered = filterByTeams(d.botPrs);
  var draftFiltered = filterByTeams(d.draftPrs);

  var stats = computeStats(d);

  var pendingIcon = '<svg viewBox="0 0 16 16"><path fill="var(--color-fg-muted)" d="M8 2a6 6 0 1 1 0 12A6 6 0 0 1 8 2Zm.75 3.25a.75.75 0 0 0-1.5 0v2.992l-1.354.902a.75.75 0 0 0 .832 1.248l1.75-1.166a.75.75 0 0 0 .272-.576Z"></path></svg>';
  var approvedIcon = '<svg viewBox="0 0 16 16"><path fill="var(--color-success-fg)" d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"></path></svg>';
  var myPrIcon = '<svg viewBox="0 0 16 16"><path fill="var(--color-done-fg)" d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25Zm5.677-.177L9.573.677A.25.25 0 0 1 10 .854V2.5h1A2.5 2.5 0 0 1 13.5 5v5.628a2.251 2.251 0 1 1-1.5 0V5a1 1 0 0 0-1-1h-1v1.646a.25.25 0 0 1-.427.177L7.177 3.427a.25.25 0 0 1 0-.354ZM3.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm0 9.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm8.25.75a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0Z"></path></svg>';

  var app = document.getElementById('app');
  app.innerHTML =
    '<div class="Subhead">' +
      '<h1 class="Subhead-heading">Pull Request Dashboard</h1>' +
      '<span class="Subhead-description">Generated <span id="last-refreshed"></span>' +
      ' <button class="refresh-btn" data-tooltip="Auto-refreshes every ' + (DATA.autoRefreshMinutes || 15) + ' min" onclick="location.reload()"><svg viewBox="0 0 16 16" width="14" height="14"><path fill="currentColor" d="M1.705 8.005a.75.75 0 0 1 .834.656 5.5 5.5 0 0 0 9.592 2.97l-1.204-1.204a.25.25 0 0 1 .177-.427h3.646a.25.25 0 0 1 .25.25v3.646a.25.25 0 0 1-.427.177l-1.07-1.07A7.002 7.002 0 0 1 1.05 8.84a.75.75 0 0 1 .656-.834ZM8 2.5a5.487 5.487 0 0 0-4.131 1.869l1.204 1.204A.25.25 0 0 1 4.896 6H1.25A.25.25 0 0 1 1 5.75V2.104a.25.25 0 0 1 .427-.177l1.07 1.07A7.002 7.002 0 0 1 14.95 7.16a.75.75 0 0 1-1.49.178A5.5 5.5 0 0 0 8 2.5Z"></path></svg></button>' +
      '</span>' +
    '</div>' +

    renderStats(stats) +
    renderFilterBar(d) +

    renderSectionGroup('Pending Reviews', pendingIcon, [
      { id: 'pending-direct', label: 'Direct review requests', prs: d.reviewDirect, countClass: 'count-red', highlightStale: true },
      { id: 'pending-team', label: 'Team review requests', prs: reviewTeamFiltered, countClass: 'count-blue', highlightStale: true },
      { id: 'pending-cr', label: 'Changes requested', prs: d.reviewChangesReq, countClass: 'count-orange', highlightStale: true },
      { id: 'pending-bot', label: 'Bot PRs', prs: botFiltered, countClass: 'count-purple', highlightStale: false },
      { id: 'pending-draft', label: 'Drafts', prs: draftFiltered, countClass: 'count-gray', highlightStale: false }
    ]) +

    renderSectionGroup('Approved Reviews', approvedIcon, [
      { id: 'approved-me', label: 'Approved by me', prs: d.approvedMe, countClass: 'count-green', highlightStale: false },
      { id: 'approved-team', label: 'Approved by team', prs: approvedTeamFiltered, countClass: 'count-green', highlightStale: false }
    ]) +

    renderSectionGroup('My Pull Requests', myPrIcon, [
      { id: 'my-approved', label: 'Approved', prs: d.myApproved, countClass: 'count-green', highlightStale: false },
      { id: 'my-cr', label: 'Changes requested', prs: d.myChangesReq, countClass: 'count-orange', highlightStale: false },
      { id: 'my-open', label: 'Open', prs: d.myOpen, countClass: 'count-blue', highlightStale: false },
      { id: 'my-drafts', label: 'Drafts', prs: d.myDraft, countClass: 'count-gray', highlightStale: false }
    ]) +

    '<div class="footer">' +
      'Generated by <a href="https://github.com/bajaj6/gh-pr-dashboard">gh pr-dashboard</a>' +
      ' &middot; ' + escapeHtml(DATA.user) + ' &middot; ' + new Date(DATA.generated).toLocaleString() +
    '</div>';

  // Populate header user
  var headerUser = document.getElementById('headerUser');
  if (headerUser) {
    headerUser.innerHTML = '<img src="https://github.com/' + escapeHtml(DATA.user) + '.png?size=40" alt="' + escapeHtml(DATA.user) + '"><span>' + escapeHtml(DATA.user) + '</span>';
  }
}

// ============================================================
// Global actions
// ============================================================
function toggleSection(id) {
  CollapsedState.toggle(id);
  var el = document.querySelector('[data-section-id="' + id + '"]');
  if (el) el.classList.toggle('collapsed');
}

// ============================================================
// Initialize
// ============================================================
CollapsedState.init();
ThemeManager.init();
render();

// Live-update "Last refreshed" display
function updateRefreshedTime() {
  var el = document.getElementById('last-refreshed');
  if (el) el.textContent = timeAgo(DATA.generated) + ' ago';
}
updateRefreshedTime();
setInterval(updateRefreshedTime, 30000);

// Auto-refresh page (configurable, 0 = disabled)
var AUTO_REFRESH_MINUTES = DATA.autoRefreshMinutes || 15;
if (AUTO_REFRESH_MINUTES > 0) {
  setInterval(function() { location.reload(); }, AUTO_REFRESH_MINUTES * 60 * 1000);
}
