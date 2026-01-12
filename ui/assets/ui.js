const UI = {
  nsSel: () => document.getElementById("namespaceSelect"),
  title: () => document.getElementById("title"),
  countInfo: () => document.getElementById("countInfo"),
  head: () => document.getElementById("tableHead"),
  body: () => document.getElementById("tableBody"),
  search: () => document.getElementById("searchInput"),
  drawer: () => document.getElementById("drawer"),
  drawerTitle: () => document.getElementById("drawerTitle"),
  drawerSubtitle: () => document.getElementById("drawerSubtitle"),
  tabContent: () => document.getElementById("tabContent"),
  tabsContainer: () => document.getElementById("tabsContainer"),
  mainLayout: () => document.getElementById("mainLayout"),
};

function openDrawer() {
  UI.mainLayout().classList.add("with-drawer");
  UI.drawer().classList.remove("closed");
}

function closeDrawer() {
  UI.mainLayout().classList.remove("with-drawer");
  UI.drawer().classList.add("closed");
}