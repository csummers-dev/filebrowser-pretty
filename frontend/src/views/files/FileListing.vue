<template>
  <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
    <!-- fb-columns (v2.7): the content columns float as SURFACE PANELS on the
         deeper canvas (gutter + rounded hairline cards — see listing.css).
         This makes the canvas↔surface separation actually visible; before,
         the main column was transparent over the shell canvas and the whole
         app read as one flat sheet. -->
    <div class="fb-columns flex-1 flex min-h-0 overflow-hidden">
      <!-- Primary column (V2-J): hero + scrolling listing + bottom breadcrumb
           bar, stacked. The details rail (InfoPane) sits beside it as a
           full-height sibling. -->
      <div
        class="fb-primary flex-1 flex flex-col min-w-0 min-h-0"
        :class="{
          'fb-primary--active': splitActive && panes.activePane === 'a',
        }"
        :role="splitActive ? 'region' : undefined"
        :aria-label="splitActive ? 'Primary folder pane' : undefined"
        @pointerdown.capture="onPaneAActivate"
      >
        <!-- V2-J unified hero: title (left) + adaptive control cluster (right).
             Replaces the dissolved top header bar; always rendered so the
             controls never flash out during a folder load. The title block is
             also the parent-folder drop / spring-load target (F2).
             In SPLIT view this hero is swapped for the compact header below so
             both panes read consistently (the breadcrumb moves to the bottom
             bar); non-split keeps the full hero unchanged. -->
        <div v-if="!splitActive" class="fb-hero">
          <!-- Mobile: open the sidebar drawer (no top bar to host it). -->
          <button
            type="button"
            class="fb-hero__menu hidden max-md:inline-flex"
            aria-label="Open navigation menu"
            title="Open menu"
            @click="mobileNav.open"
          >
            <Icon name="menu" :size="18" />
          </button>

          <!-- Hero lead (Calm Minimal #1 + #3): the breadcrumb is now the
               primary location nav — its last crumb is the current folder — with
               the favorite star and a compact item count beside it. The former
               FOLDER eyebrow + jumbo title + meta line were condensed away (the
               breadcrumb shows where you are) and the bottom breadcrumb bar is
               dropped in single-pane. Drag-to-parent + spring-up live on the
               breadcrumb crumbs themselves now, so the section-title drop
               handlers aren't needed here. Renaming the current folder swaps the
               crumb row for an inline input. -->
          <div
            v-if="!layoutStore.loading && fileStore.req && fileStore.req.isDir"
            class="fb-hero__lead min-w-0"
          >
            <input
              v-if="isRenamingCurrentFolder"
              ref="folderRenameInputEl"
              v-model.trim="folderRenameValue"
              class="folder-rename-input text-[15px] font-semibold text-ink-1 leading-tight truncate"
              type="text"
              autocomplete="off"
              spellcheck="false"
              @keydown.enter.prevent="submitFolderRename"
              @keydown.esc.prevent="cancelFolderRename"
              @blur="onFolderRenameBlur"
            />
            <div v-else class="fb-hero__crumbs">
              <breadcrumbs base="/files" />
              <button
                v-if="fileStore.req?.isDir && currentFolderPath"
                type="button"
                class="current-fav-btn"
                :class="{ 'current-fav-btn--active': currentFolderFavorited }"
                :title="
                  currentFolderFavorited
                    ? 'Remove from Favorites'
                    : 'Add to Favorites'
                "
                :aria-label="
                  currentFolderFavorited
                    ? 'Remove from Favorites'
                    : 'Add to Favorites'
                "
                :aria-pressed="currentFolderFavorited"
                @click.stop="onCurrentFolderFavToggle"
              >
                <Icon
                  name="star"
                  :size="13"
                  :stroke-width="currentFolderFavorited ? 0 : 2"
                  :fill="currentFolderFavorited ? 'currentColor' : 'none'"
                />
              </button>
              <span class="fb-hero__count" :title="folderMeta">{{
                folderCountLabel
              }}</span>
            </div>
          </div>
          <div v-else class="fb-hero__title"></div>

          <!-- Hero right column: control cluster on top, search field stacked
               directly beneath it. Keeping search in this right-hand column makes
               it sit FLUSH under the toolbar — a separate full-width row below the
               hero dropped beneath the multi-line title block on the left, leaving
               a big vertical gap. -->
          <div class="fb-hero__right">
            <!-- Control cluster — single adaptive row. Share + Download collapse
                 into the ⋯ More menu to save space (V2-J). -->
            <div class="fb-hero__cluster">
              <!-- One consolidated View control (v2.7.2 header declutter):
                   the 3-button layout switcher + the split toggle collapse
                   into a single popover. The trigger wears the ACTIVE mode's
                   icon so the current layout stays glanceable; the menu holds
                   List / Grid / Gallery plus Split view. -->
              <button
                class="h-7 w-7 rounded-md border border-line bg-surface hover:bg-elevated inline-flex items-center justify-center transition"
                @click.stop="openViewMenu"
                :title="`View: ${viewModeLabel}`"
                :aria-label="`View options — current: ${viewModeLabel}`"
                aria-haspopup="menu"
              >
                <Icon
                  :name="viewModeIcon"
                  :size="14"
                  class="text-[var(--color-ink-2)]"
                />
              </button>
              <!-- One consolidated Sort control: opens a popover with the sort
                   field, direction (Ascending/Descending), and the secondary
                   "Then by" criterion — replacing the former two-button field +
                   direction pair. Icon-only at all widths; the active field +
                   direction stay legible via the title tooltip. -->
              <button
                class="h-7 w-7 rounded-md border border-line bg-surface hover:bg-elevated inline-flex items-center justify-center transition"
                @click.stop="openSortMenu"
                :title="`Sort: ${sortLabel} · ${sortDirLabel}`"
                :aria-label="`Sort: ${sortLabel}, ${sortDirLabel}`"
              >
                <Icon
                  name="arrow-up-down"
                  :size="14"
                  class="text-[var(--color-ink-2)]"
                />
              </button>
              <button
                v-if="headerButtons.upload && isTouchDevice"
                class="h-7 w-7 rounded-md border border-line bg-surface hover:bg-elevated text-ink-2 inline-flex items-center justify-center transition"
                @click="photoInput?.click()"
                title="Add photos or videos"
                aria-label="Add photos or videos"
              >
                <Icon name="image-plus" :size="14" />
              </button>
              <input
                ref="photoInput"
                style="display: none"
                type="file"
                accept="image/*,video/*"
                multiple
                @change="uploadInput($event)"
              />
              <!-- Icon-only at all widths (label dropped to save header room;
                   the gradient fill keeps it reading as the primary action). -->
              <button
                v-if="headerButtons.upload"
                class="h-7 w-7 rounded-md btn-accent-gradient text-white flex items-center justify-center shadow-sm transition"
                @click="uploadFunc"
                :title="t('buttons.upload')"
                :aria-label="t('buttons.upload')"
              >
                <Icon name="upload" :size="14" />
              </button>
              <button
                class="w-7 h-7 rounded-md border border-line bg-surface hover:bg-elevated text-ink-2 flex items-center justify-center transition"
                title="More"
                aria-label="More"
                @click.stop="showSectionMore"
              >
                <Icon name="ellipsis" :size="14" />
              </button>
            </div>
            <!-- Search sits directly beneath the toolbar cluster (single-pane —
                 the whole hero is gated !splitActive). On its own line it expands
                 leftward freely without crowding the buttons; right-aligned under
                 the cluster via the column's align-items. -->
            <search />
          </div>
        </div>
        <!-- Split-only compact header (Request #2): in split view both panes
             share pane B's toolbar chrome via the global .compare-head /
             .compare-btn classes, so the headers are identical. The breadcrumb
             lives in the bottom bar; a spacer right-aligns the action cluster.
             Button set matches pane B exactly (parent · sort · sort-dir · new
             folder · upload · ⋯ · close). Search stays available via ⌘K. -->
        <header
          v-if="splitActive"
          class="compare-head"
          :class="{ 'compare-head--drop': sectionDropActive }"
          @dragenter="onSectionDragEnter"
          @dragover="onSectionDragOver"
          @dragleave="onSectionDragLeave"
          @drop="onSectionDrop"
        >
          <div class="compare-head__info" :title="folderTitle">
            <Icon
              name="folder"
              :size="15"
              class="compare-head__icon text-[var(--color-ink-2)]"
            />
            <span class="compare-head__name headline-gradient">{{
              folderTitle
            }}</span>
            <span v-if="splitHeaderCount > 0" class="compare-head__count">
              {{ splitHeaderCount }}
              {{ splitHeaderCount === 1 ? "item" : "items" }}
            </span>
          </div>
          <div class="compare-head__actions">
            <button
              type="button"
              class="compare-head__parent"
              :disabled="!parentFolderUrl"
              title="Parent folder"
              aria-label="Parent folder"
              @click="goToParentFolder"
            >
              <Icon name="arrow-up" :size="15" :stroke-width="2.2" />
            </button>
            <button
              type="button"
              class="compare-btn compare-btn--icon"
              :title="`Sort: ${sortLabel} · ${sortDirLabel}`"
              :aria-label="`Sort: ${sortLabel}, ${sortDirLabel}`"
              @click.stop="openSortMenu"
            >
              <Icon
                name="arrow-up-down"
                :size="14"
                class="text-[var(--color-ink-2)]"
              />
            </button>
            <span class="compare-head__sep" aria-hidden="true"></span>
            <button
              v-if="headerButtons.upload"
              type="button"
              class="compare-btn"
              title="New folder"
              aria-label="New folder"
              @click="layoutStore.showHover('newDir')"
            >
              <Icon
                name="folder-plus"
                :size="14"
                class="text-[var(--color-ink-2)]"
              />
            </button>
            <button
              v-if="headerButtons.upload"
              type="button"
              class="compare-btn"
              :title="t('buttons.upload')"
              :aria-label="t('buttons.upload')"
              @click="uploadFunc"
            >
              <Icon
                name="upload"
                :size="14"
                class="text-[var(--color-ink-2)]"
              />
            </button>
            <button
              type="button"
              class="compare-btn"
              title="More actions"
              aria-label="More actions"
              @click.stop="showSectionMore"
            >
              <Icon name="ellipsis" :size="15" />
            </button>
            <button
              type="button"
              class="compare-btn compare-btn--close"
              title="Close this pane"
              aria-label="Close this pane"
              @click="closeSplitToB"
            >
              <Icon name="x" :size="15" />
            </button>
          </div>
        </header>
        <!-- Sort dropdown — hoisted out of the hero cluster so it renders in
             BOTH the hero and the split compact header. -->
        <context-menu
          :show="sortMenuShow"
          :pos="sortMenuPos"
          :items="sortMenuItems"
          @hide="sortMenuShow = false"
        />
        <!-- View popover (v2.7.2): layout + split, one trigger. -->
        <context-menu
          :show="viewMenuShow"
          :pos="viewMenuPos"
          :items="viewMenuItems"
          @hide="viewMenuShow = false"
        />
        <!-- Type filter chips (v2.7): client-side filter for the current
             folder's FILES (folders always stay visible). Rendered only when
             the folder actually mixes ≥2 categories, and not in split (the
             half-width panes have no room for a chip row). -->
        <div
          v-if="showTypeChips"
          class="fb-type-chips"
          role="group"
          aria-label="Filter files by type"
        >
          <button
            type="button"
            class="fb-type-chip"
            :class="{ 'fb-type-chip--on': typeFilter === 'all' }"
            :aria-pressed="typeFilter === 'all'"
            @click="setTypeFilter('all')"
          >
            All
          </button>
          <button
            v-for="opt in typeChipList"
            :key="opt.key"
            type="button"
            class="fb-type-chip"
            :class="{ 'fb-type-chip--on': typeFilter === opt.key }"
            :aria-pressed="typeFilter === opt.key"
            @click="setTypeFilter(opt.key)"
          >
            {{ opt.label }}
            <span class="fb-type-chip__n">{{ opt.count }}</span>
          </button>
        </div>
        <section
          ref="scrollSection"
          class="flex-1 flex flex-col min-w-0 min-h-0 overflow-y-auto ptr-host"
          :class="{
            'scroll-section--virtual': isListVirtual,
            'scroll-section--drop': paneADropActive,
          }"
          @dragenter="onPaneADragEnter"
          @dragover="onPaneADragOver"
          @dragleave="onPaneADragLeave"
          @drop="onPaneADrop"
        >
          <!-- S7-1: pull-to-refresh indicator. Only visible during a pull
             or while refreshing; pinned at the visible top (scrollTop is
             held at 0 throughout the gesture, so absolute top:0 = top of
             the viewport). -->
          <div
            v-if="ptrActive"
            class="ptr-indicator"
            :class="{ 'ptr-indicator--pulling': ptrPulling }"
            :style="{
              transform: `translate(-50%, ${ptrOffset}px)`,
              opacity: ptrOpacity,
            }"
            aria-hidden="true"
          >
            <div class="ptr-indicator__pill">
              <Icon
                name="loader-circle"
                :size="18"
                class="ptr-indicator__spin"
                :class="{ 'ptr-indicator__spin--active': ptrRefreshing }"
                :style="
                  ptrRefreshing
                    ? undefined
                    : { transform: `rotate(${ptrRotation}deg)` }
                "
              />
            </div>
          </div>

          <div
            v-if="isMobile && fileStore.selectedCount > 0"
            id="file-selection"
            :class="{
              'file-selection-margin-bottom': fileStore.multiple,
            }"
          >
            <span>
              {{ t("prompts.filesSelected", fileStore.selectedCount) }}
            </span>
            <!-- Opt-in details toggle (mobile). The InfoPane sheet no longer
               pops on its own when an item is selected; the user expands it
               here when they actually want metadata. -->
            <button
              type="button"
              class="action"
              title="Details"
              aria-label="Details"
              @click="layoutStore.mobileDetailsOpen = true"
            >
              <Icon name="info" :size="18" />
            </button>
            <action
              v-if="headerButtons.share"
              icon="share"
              :label="t('buttons.share')"
              show="share"
            />
            <action
              v-if="headerButtons.rename"
              icon="pencil"
              :label="t('buttons.rename')"
              show="rename"
            />
            <action
              v-if="headerButtons.copy"
              icon="copy"
              :label="t('buttons.copyFile')"
              show="copy"
            />
            <action
              v-if="headerButtons.move"
              icon="forward"
              :label="t('buttons.moveFile')"
              show="move"
            />
            <action
              v-if="headerButtons.delete"
              icon="trash-2"
              :label="t('buttons.delete')"
              show="delete"
            />
          </div>

          <!-- Section title "More" dropdown.
             Migrated to the items-based ContextMenu API (v1.3.0 S1-3) —
             keyboard nav, type-ahead, separators all come along for
             free. The action array below is built in the script. -->
          <context-menu
            :show="sectionMoreShow"
            :pos="sectionMorePos"
            :items="sectionMoreItems"
            @hide="hideSectionMore"
          />

          <!-- Loading: skeleton rows matching the current view mode. Mirrors
         the real ListingItem layout so the page doesn't reflow on load. -->
          <ListingSkeleton
            v-if="layoutStore.loading"
            :mode="viewMode === 'list' ? 'list' : 'mosaic'"
            :count="viewMode === 'list' ? 10 : 12"
          />
          <template v-else>
            <!-- Empty state vs. listing chrome. Critical detail (H6):
               when the user is mid-creation (`inlineNewKind` is set)
               we MUST render the listing branch even on an empty
               folder — otherwise the InlineNewItem (which only lives
               inside #listing) has nowhere to mount and clicking
               "New Folder" silently no-ops. The listing's header +
               inline input form a clean "creating in empty folder"
               surface; EmptyState comes back when the prompt closes. -->
            <div
              v-if="
                (fileStore.req?.numDirs ?? 0) +
                  (fileStore.req?.numFiles ?? 0) ==
                  0 && !inlineNewKind
              "
            >
              <EmptyState
                icon="folder-open"
                :title="
                  folderTitle ? `${folderTitle} is empty` : t('files.lonely')
                "
                hint="Drop files here or use the Upload button to add content."
              >
                <button
                  v-if="headerButtons.upload"
                  type="button"
                  class="empty-cta"
                  @click="uploadFunc"
                >
                  <Icon name="upload" :size="13" />
                  <span>{{ t("buttons.upload") }}</span>
                </button>
              </EmptyState>
              <input
                style="display: none"
                type="file"
                id="upload-input"
                @change="uploadInput($event)"
                multiple
              />
              <input
                style="display: none"
                type="file"
                id="upload-folder-input"
                @change="uploadInput($event)"
                webkitdirectory
                multiple
              />
            </div>
            <div
              v-else
              id="listing"
              ref="listing"
              class="file-icons"
              data-clear-on-click="true"
              :data-drop-url="currentFolderUrl || undefined"
              :class="[
                effectiveViewMode,
                {
                  'listing--virtual': isListVirtual,
                  'is-touch': isTouchDevice,
                },
              ]"
              @click="handleEmptyAreaClick"
              @contextmenu="onListingContextMenu"
              @mousedown.capture="recordListingPressOrigin"
              @mousedown="dragSelect.onMouseDown"
            >
              <div>
                <div class="item header">
                  <div class="item__select">
                    <div
                      :class="[
                        'item__checkbox',
                        allSelected && 'item__checkbox--checked',
                      ]"
                      role="checkbox"
                      :aria-checked="allSelected"
                      tabindex="0"
                      @click.stop="toggleSelectAll"
                      :title="allSelected ? 'Deselect all' : 'Select all'"
                    >
                      <Icon
                        v-if="allSelected"
                        name="check"
                        :size="11"
                        :stroke-width="3.5"
                        style="color: white"
                      />
                    </div>
                  </div>
                  <p
                    :class="{ active: nameSorted }"
                    class="name"
                    role="button"
                    tabindex="0"
                    @click="sort('name')"
                    :title="t('files.sortByName')"
                    :aria-label="t('files.sortByName')"
                  >
                    <span>{{ t("files.name") }}</span>
                    <Icon :name="nameIcon" />
                  </p>
                  <p
                    :class="{ active: modifiedSorted }"
                    class="modified"
                    role="button"
                    tabindex="0"
                    @click="sort('modified')"
                    :title="t('files.sortByLastModified')"
                    :aria-label="t('files.sortByLastModified')"
                  >
                    <span>Modified</span>
                    <Icon :name="modifiedIcon" />
                  </p>
                  <p
                    :class="{ active: sizeSorted }"
                    class="size"
                    role="button"
                    tabindex="0"
                    @click="sort('size')"
                    :title="t('files.sortBySize')"
                    :aria-label="t('files.sortBySize')"
                  >
                    <span>{{ t("files.size") }}</span>
                    <Icon :name="sizeIcon" />
                  </p>
                  <div class="header__actions"></div>
                </div>
              </div>

              <!-- Inline new folder / new file row (Stage 8). Replaces the legacy
             modal: appears at the top of the listing with a focused input. -->
              <div v-if="inlineNewKind" @click.stop>
                <InlineNewItem :kind="inlineNewKind" />
              </div>

              <!-- v1.3 S6-1: the LIST view is virtualized with
                 RecycleScroller (fixed 44px rows) so a 10k-file folder
                 only ever mounts the visible window + a small buffer.
                 dirs → divider → files are flattened into `listRows`;
                 selection/drag/rename all stay index-based so they keep
                 working through recycling. Grid + gallery keep the
                 incremental `showLimit` windowing in the v-else branch —
                 their tiles wrap to variable heights, so list-first per
                 the locked Stage 6 plan. -->
              <!-- `:key` on typeFilter: shrinking `items` IN PLACE leaves the
                 RecycleScroller's pooled row views visible at their old
                 offsets (the wrapper resizes but stale views linger — v2.7
                 chips bug, verified live). Nothing else shrinks the list
                 without swapping `req`, so re-creating the scroller per
                 filter state is the surgical fix; scroll resets, which a
                 filter change wants anyway. -->
              <RecycleScroller
                v-if="isListVirtual"
                ref="listScroller"
                :key="'vlist-' + typeFilter"
                class="listing-virtual"
                :items="listRows"
                :item-size="null"
                :min-item-size="40"
                key-field="id"
                :buffer="320"
                data-clear-on-click="true"
                v-slot="{ item: row }"
              >
                <div
                  v-if="row.divider"
                  class="folder-file-divider"
                  data-clear-on-click="true"
                ></div>
                <item
                  v-else-if="row.item"
                  v-bind:index="row.item.index"
                  v-bind:name="row.item.name"
                  v-bind:isDir="row.item.isDir"
                  v-bind:url="row.item.url"
                  v-bind:modified="row.item.modified"
                  v-bind:type="row.item.type"
                  v-bind:size="row.item.size"
                  v-bind:path="row.item.path"
                  @dropAlongside="onItemDropAlongside"
                  @rowPointerDown="onItemPointerDown"
                >
                </item>
              </RecycleScroller>

              <template v-else>
                <!-- CH-1: the folders + files grids are windowed — only the
                   tiles near the viewport mount. The top/bottom spacer
                   blocks stand in for the rows scrolled out of view so the
                   scroll height + position stay exact. Spacers are block
                   siblings of the grid (not grid cells), so there's no gap
                   math to get wrong, and each section's total height stays
                   constant as you scroll. -->
                <div
                  v-if="(fileStore.req?.numDirs ?? 0) > 0"
                  ref="dirsSectionRef"
                  class="listing-section-pad"
                  data-clear-on-click="true"
                >
                  <div
                    class="grid-pad"
                    data-clear-on-click="true"
                    :style="{ height: dirsWin.topPad + 'px' }"
                  ></div>
                  <div
                    class="listing-section listing-section--dirs"
                    data-clear-on-click="true"
                  >
                    <item
                      v-for="item in dirs"
                      :key="base64(item.name)"
                      v-bind:index="item.index"
                      v-bind:name="item.name"
                      v-bind:isDir="item.isDir"
                      v-bind:url="item.url"
                      v-bind:modified="item.modified"
                      v-bind:type="item.type"
                      v-bind:size="item.size"
                      v-bind:path="item.path"
                      @dropAlongside="onItemDropAlongside"
                      @rowPointerDown="onItemPointerDown"
                    >
                    </item>
                  </div>
                  <div
                    class="grid-pad"
                    data-clear-on-click="true"
                    :style="{ height: dirsWin.botPad + 'px' }"
                  ></div>
                </div>

                <div
                  v-if="
                    (fileStore.req?.numDirs ?? 0) > 0 &&
                    (fileStore.req?.numFiles ?? 0) > 0
                  "
                  class="folder-file-divider"
                  data-clear-on-click="true"
                ></div>

                <div
                  v-if="(fileStore.req?.numFiles ?? 0) > 0"
                  ref="filesSectionRef"
                  class="listing-section-pad"
                  data-clear-on-click="true"
                >
                  <div
                    class="grid-pad"
                    data-clear-on-click="true"
                    :style="{ height: filesWin.topPad + 'px' }"
                  ></div>
                  <div class="listing-section" data-clear-on-click="true">
                    <item
                      v-for="item in files"
                      :key="base64(item.name)"
                      v-bind:index="item.index"
                      v-bind:name="item.name"
                      v-bind:isDir="item.isDir"
                      v-bind:url="item.url"
                      v-bind:modified="item.modified"
                      v-bind:type="item.type"
                      v-bind:size="item.size"
                      v-bind:path="item.path"
                      @dropAlongside="onItemDropAlongside"
                      @rowPointerDown="onItemPointerDown"
                    >
                    </item>
                  </div>
                  <div
                    class="grid-pad"
                    data-clear-on-click="true"
                    :style="{ height: filesWin.botPad + 'px' }"
                  ></div>
                </div>
              </template>

              <!-- S4-1: items-based context menu. The same `<context-menu>`
                 instance serves both right-click on a row (rowMenuItems)
                 and right-click on empty listing space (background-
                 MenuItems) — `onListingContextMenu` decides which to
                 hand it based on event.target. ContextMenu handles
                 keyboard nav, type-ahead, smart positioning. -->
              <context-menu
                :show="isContextMenuVisible"
                :pos="contextMenuPos"
                :items="contextMenuItems"
                @hide="hideContextMenu"
              />

              <input
                style="display: none"
                type="file"
                id="upload-input"
                @change="uploadInput($event)"
                multiple
              />
              <input
                style="display: none"
                type="file"
                id="upload-folder-input"
                @change="uploadInput($event)"
                webkitdirectory
                multiple
              />
            </div>
          </template>
        </section>
        <!-- Bulk-selection pill. A direct child of `.fb-primary` (NOT the
             scroll section) so it can be `position: absolute` within the pane's
             column — keeping it over pane A only in split view, instead of
             viewport-centred across both panes. Visibility is driven by
             `showBulkPill`; it slides up from below when ≥2 items are selected. -->
        <div
          :class="{
            active: showBulkPill,
            'multiple-selection--compact': splitActive,
          }"
          id="multiple-selection"
          @click.stop
        >
          <span class="multiple-selection__count">
            {{ fileStore.selectedCount }} selected
          </span>

          <div class="multiple-selection__divider"></div>

          <button
            v-if="headerButtons.download"
            @click="download"
            :title="t('buttons.download')"
          >
            <!-- Calm Minimal: action icons are a uniform soft white on the dark
                 pill (was per-action bright hues), matching the white labels. -->
            <Icon name="download" :size="14" class="text-white/80" />
            <span>{{ t("buttons.download") }}</span>
          </button>
          <button
            v-if="headerButtons.copy"
            @click="layoutStore.showHover('copy')"
            :title="t('buttons.copyFiles')"
          >
            <Icon name="copy" :size="14" class="text-white/80" />
            <span>{{ t("buttons.copyFiles") }}</span>
          </button>
          <button
            v-if="headerButtons.move"
            @click="layoutStore.showHover('move')"
            :title="t('buttons.moveFiles')"
          >
            <Icon name="forward" :size="14" class="text-white/80" />
            <span>{{ t("buttons.moveFiles") }}</span>
          </button>
          <!-- Extract shows only when EVERY selected item is an archive
               (headerButtons.extract); each extracts into its own subfolder. -->
          <button
            v-if="headerButtons.extract"
            @click="layoutStore.showHover('extract')"
            title="Extract archives"
          >
            <Icon name="package-open" :size="14" class="text-white/80" />
            <span>Extract</span>
          </button>
          <!-- v1.3 S4-2: Bulk rename pill button. Only renders for multi-select
               (single-select uses inline rename in ListingItem). Gated on
               perm.rename. -->
          <button
            v-if="fileStore.selectedCount > 1 && authStore.user?.perm.rename"
            @click="bulkRename.open"
            title="Bulk rename"
          >
            <Icon name="pencil" :size="14" class="text-white/80" />
            <span>Rename</span>
          </button>
          <!-- 1.6.0: batch-edit ID3/Vorbis tags across the audio files in the
               selection. Only fields changed in the editor are applied to all. -->
          <button
            v-if="canBulkEditTags"
            @click="layoutStore.showHover('audio-tags')"
            :title="`Edit tags on ${bulkAudioCount} audio files`"
          >
            <Icon name="music" :size="14" class="text-white/80" />
            <span>Edit tags</span>
          </button>
          <!-- 2.4.0 Stage 5 / K: bulk add/remove user tags across the whole
               selection (tri-state for tags on some-but-not-all). -->
          <button
            v-if="fileStore.selectedCount > 1"
            @click="openBulkTags"
            :title="`Tag ${fileStore.selectedCount} items`"
          >
            <Icon name="tag" :size="14" class="text-white/80" />
            <span>Tags</span>
          </button>
          <button
            v-if="headerButtons.delete"
            @click="layoutStore.showHover('delete')"
            class="multiple-selection__danger"
            :title="t('buttons.delete')"
          >
            <Icon name="trash-2" :size="14" class="text-white/80" />
            <span>{{ t("buttons.delete") }}</span>
          </button>

          <div class="multiple-selection__divider"></div>

          <button
            @click="clearSelection"
            class="multiple-selection__close"
            :title="t('buttons.clear')"
            :aria-label="t('buttons.clear')"
          >
            <Icon name="x" :size="14" />
          </button>
        </div>
        <!-- Bottom breadcrumb path — split mode only. In single-pane the
             breadcrumb now leads the hero at the TOP (Calm Minimal #3), so a
             bottom bar would duplicate it; in split mode pane A keeps its path
             here (its compact header has no room) and pane B mirrors it with its
             own footer crumbs. Listing pages only. -->
        <div v-if="splitActive" class="fb-breadcrumb-bar">
          <div class="fb-breadcrumb-bar__crumbs">
            <breadcrumbs base="/files" />
          </div>
        </div>
      </div>
      <!-- Dual-pane second pane (2.5.0). Sits beside the primary column; the
           details rail is hidden while split (the two listings need the room). -->
      <ComparePane v-if="splitActive" />
      <InfoPane v-else />
    </div>

    <!-- Space-bar quick preview (Quick Look). One instance serves both panes;
         it teleports to <body> and reads the useQuickPeek singleton. -->
    <QuickPeek />

    <!-- Delete confirmation (Stage 8; trash-aware since 2.4.0 Stage 2). -->
    <ConfirmDialog
      :open="confirmOpen"
      :title="confirmTitle"
      :message="confirmMessage"
      :confirm-label="pendingPermanent ? 'Delete forever' : 'Move to Trash'"
      cancel-label="Cancel"
      destructive
      @confirm="onDeleteConfirm"
      @cancel="onDeleteCancel"
    />

    <!-- Move / Copy slide-over (Stage 8). One component, two modes. -->
    <MoveCopyPanel
      :open="moveCopyOpen"
      :mode="moveCopyMode"
      :override="moveCopyOverride"
      @cancel="closeMoveCopy"
      @done="closeMoveCopy"
    />

    <!-- Bulk rename slide-over (v1.3 S4-2). Pattern / find-replace
         modes with live preview + conflict highlighting; client-loops
         api.move with continue-on-error semantics. Open-state lives in
         the useBulkRename singleton so the command palette can trigger
         it without prop-drilling through FileListing. -->
    <BulkRenamePanel
      :open="bulkRename.isOpen.value"
      @cancel="bulkRename.close"
      @done="bulkRename.close"
    />

    <!-- 2.4.0 Stage 5 / K: bulk tag picker for a multi-selection. Its own
         TagPickerSheet instance (the single-file one lives in InfoPane), bound
         to the useBulkTagPicker singleton + a `paths` array → tri-state. -->
    <TagPickerSheet
      :open="bulkTagPicker.isOpen.value"
      :paths="bulkTagPicker.paths.value"
      @cancel="bulkTagPicker.close"
      @saved="bulkTagPicker.close"
    />

    <!-- Share slide-over (Stage 8). The optional `override` lets the second
         pane (ComparePane) target its own item via showHover props. -->
    <SharePanel
      :open="shareOpen"
      :override="layoutStore.currentPrompt?.props?.override"
      @cancel="closeShare"
    />

    <!-- Extract zip slide-over (PR #5746 Phase B). -->
    <ExtractPanel
      :open="extractOpen"
      :override="layoutStore.currentPrompt?.props?.override"
      @cancel="closeExtract"
      @done="closeExtract"
    />

    <!-- Drag-select lasso rectangle (v1.3 S4-3). Teleported to body so
         fixed-position viewport coords aren't clipped by the listing's
         overflow. Only present while a marquee is active in grid /
         gallery. -->
    <Teleport to="body">
      <div
        v-if="dragSelect.lasso.value"
        class="drag-lasso"
        :style="{
          left: dragSelect.lasso.value.x + 'px',
          top: dragSelect.lasso.value.y + 'px',
          width: dragSelect.lasso.value.w + 'px',
          height: dragSelect.lasso.value.h + 'px',
        }"
      ></div>
    </Teleport>

    <!-- Image hover-preview overlay (v1.3 S5-9). Single instance driven
         by the useImageHoverPreview singleton; rows schedule/cancel it.
         Self-teleports to body + size-caps the image. -->
    <ImageHoverPreview />
  </div>
</template>

<script setup lang="ts">
import Icon from "@/components/Icon.vue";
import { useAuthStore } from "@/stores/auth";
import { useClipboardStore } from "@/stores/clipboard";
import { useFileStore } from "@/stores/file";
import { usePanesStore } from "@/stores/panes";
import { useLayoutStore } from "@/stores/layout";
import { useTagsStore } from "@/stores/tags";
import { useFavorites } from "@/composables/useFavorites";
import { useRootLabel } from "@/composables/useRootLabel";
import { usePreferences } from "@/composables/usePreferences";
import { useFolderScrollMemory } from "@/composables/useFolderScrollMemory";
import { useFavoriteTitleDialog } from "@/composables/useFavoriteTitleDialog";
import { useBulkRename } from "@/composables/useBulkRename";
import { useDragSelect } from "@/composables/useDragSelect";
import { useListingGrid } from "@/composables/useListingGrid";
import { useListingNavigation } from "@/composables/useListingNavigation";
import { useTouchDevice } from "@/composables/useTouchDevice";
import { usePullToRefresh } from "@/composables/usePullToRefresh";
import { copy } from "@/utils/clipboard";
import { sortListing } from "@/utils/secondarySort";
import { useFolderSizes } from "@/composables/useFolderSizes";

import { users, files as api, trash as trashApi } from "@/api";
import { enableExec, unzipEnabled } from "@/utils/constants";
import { isExtractable } from "@/utils/archive";
import { isAudioTaggable } from "@/utils/audio";
import { TypeaheadSession } from "@/utils/typeahead";
import { moveDragBadge, endDragBadge } from "@/utils/dragCopyMoveBadge";
import * as upload from "@/utils/upload";
import { throttle } from "lodash-es";
import { Base64 } from "js-base64";
import { timeAgo } from "@/utils/relativeTime";
// v1.3 S6-1: fixed-size list virtualization for huge folders.
import { RecycleScroller } from "vue-virtual-scroller";
import "vue-virtual-scroller/dist/vue-virtual-scroller.css";

import { useMobileNav } from "@/composables/useMobileNav";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Action from "@/components/header/Action.vue";
import Search from "@/components/Search.vue";
import Item from "@/components/files/ListingItem.vue";
import InfoPane from "@/components/files/InfoPane.vue";
import ComparePane from "@/components/files/ComparePane.vue";
import QuickPeek from "@/components/files/QuickPeek.vue";
import { useQuickPeek } from "@/composables/useQuickPeek";
import {
  typeFilterCategory,
  TYPE_FILTER_LABELS,
  TYPE_FILTER_ORDER,
  type TypeFilterKey,
} from "@/utils/typeFilter";
import { pastedFileName } from "@/utils/filename";
import { pressStartedOnEmptyListing } from "@/utils/listingPress";
import ImageHoverPreview from "@/components/files/ImageHoverPreview.vue";
import InlineNewItem from "@/components/files/InlineNewItem.vue";
import ListingSkeleton from "@/components/files/ListingSkeleton.vue";
import EmptyState from "@/components/EmptyState.vue";
import UndoToast from "@/components/UndoToast.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import MoveCopyPanel from "@/components/files/MoveCopyPanel.vue";
import BulkRenamePanel from "@/components/files/BulkRenamePanel.vue";
import TagPickerSheet from "@/components/files/TagPickerSheet.vue";
import { useBulkTagPicker } from "@/composables/useBulkTagPicker";
import ExtractPanel from "@/components/files/ExtractPanel.vue";
import SharePanel from "@/components/files/SharePanel.vue";
import { useToast } from "vue-toastification";
import ContextMenu, { type MenuItem } from "@/components/ContextMenu.vue";
import { useDropTarget } from "@/composables/useDropTarget";
import { useTouchDrag } from "@/composables/useTouchDrag";
import {
  computed,
  inject,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import { useRoute, useRouter, onBeforeRouteUpdate } from "vue-router";
import url from "@/utils/url";
import { useI18n } from "vue-i18n";
import { storeToRefs } from "pinia";
import { removePrefix } from "@/api/utils";
import { startTransfer } from "@/utils/transfers";
import { resolveRowDropMode } from "@/utils/dropZone";

const dragCounter = ref<number>(0);
const width = ref<number>(window.innerWidth);
const isContextMenuVisible = ref<boolean>(false);

// ── Dual-pane / split view (2.5.0) ───────────────────────────────────
// `panes.split` is the user's choice; `splitActive` also gates on width — the
// split needs room, so below SPLIT_MIN_WIDTH (or on mobile) we collapse to a
// single pane without forgetting the preference. While active, this pane (A)
// is forced to the list layout (per the locked decision) via effectiveViewMode,
// without clobbering the user's stored grid/gallery preference.
const panes = usePanesStore();
const SPLIT_MIN_WIDTH = 880;
const splitAvailable = computed(() => width.value >= SPLIT_MIN_WIDTH);
const splitActive = computed(() => panes.split && splitAvailable.value);
const effectiveViewMode = computed(() =>
  splitActive.value ? "list" : viewMode.value
);
// Item count shown in pane A's split-only compact header (mirrors pane B).
const splitHeaderCount = computed(
  () => (fileStore.req?.numDirs ?? 0) + (fileStore.req?.numFiles ?? 0)
);
const toggleSplit = () => {
  if (panes.split) panes.closeSplit();
  else panes.openSplit(route.path.replace(/\/?$/, "/"));
};
// Interacting with the primary column makes pane A the active pane again (the
// mirror of ComparePane's pointerdown), so global surfaces target it. No-op
// when not split. Capture-phase so it fires before row handlers stop it.
const onPaneAActivate = () => {
  if (splitActive.value && panes.activePane !== "a") panes.setActive("a");
};
// Keep the active-pane target consistent with what's actually on screen: when
// the split collapses (narrowed below the min width, or toggled off) pane B is
// no longer mounted, so pane A must become active again — otherwise the
// sidebar/palette could navigate an invisible pane B.
//
// `immediate: true` is important: `panes` is a singleton that outlives this
// component, and the reset only fires on a *transition*. Without the immediate
// pass, a re-mount that lands with the split already collapsed but a stale
// `activePane === "b"` (e.g. split on + pane B clicked → leave /files → narrow
// the window → return) would never self-correct. That stuck state silently
// breaks single-pane inline rename (its `isRenaming` gate requires
// `activePane === "a"`), which a page refresh — resetting the store — "fixes".
watch(
  splitActive,
  (active) => {
    if (!active && panes.activePane !== "a") panes.setActive("a");
  },
  { immediate: true }
);
const contextMenuPos = ref<{ x: number; y: number }>({ x: 0, y: 0 });

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const clipboardStore = useClipboardStore();
const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const tagsStore = useTagsStore();
const prefs = usePreferences();
// Shared lazy folder-size cache — feeds both the Size column display and,
// when sorting by size, the folder sort order (see the `items` computed).
const folderSizes = useFolderSizes();
// V2-J: the mobile hamburger lives in the hero now (no top header bar).
const mobileNav = useMobileNav();
// Shared open-state for the favorites display-title editor, opened from the
// row context menu + the section ⋯ menu (both gated on the folder being pinned).
const favTitleDialog = useFavoriteTitleDialog();
// S4-2: shared open-state for the BulkRenamePanel so the row context
// menu, the bulk-pill button, and the command palette all trigger
// through the same flag.
const bulkRename = useBulkRename();
// 2.4.0 Stage 5 / K: bulk tag picker for the current multi-selection.
const bulkTagPicker = useBulkTagPicker();
const openBulkTags = () => {
  const paths = fileStore.selected
    .map((i) => fileStore.req?.items[i]?.url)
    .filter((u): u is string => !!u);
  if (paths.length) bulkTagPicker.open(paths);
};

// ── Inline tag chip plumbing (v1.3 S2-5) ────────────────────────────
// After each listing fetch, do ONE batched call to /api/tags/batch so
// every row can render its chips without N+1 lookups. Triggered by a
// watch on fileStore.req — fires on initial load AND every reload
// (rename, paste, upload, etc.). Failure is silent: chips just don't
// render. The store also pre-loads the user's full tag list so the
// info-pane picker has its autocomplete ready.
watch(
  () => fileStore.req,
  (req) => {
    if (!req?.isDir || !Array.isArray(req.items)) return;
    void tagsStore.ensureLoaded();
    const paths = req.items.map((i) => i.url).filter(Boolean);
    void tagsStore.loadForPaths(paths);
    // Folder sizes are fetched lazily per visible row now (ListingItem), so the
    // Size column fills in for what's on screen without prefetching every
    // subfolder up front (DP v2 R1 — that recursive walk doubled with the split).
  },
  { immediate: true }
);

const { req } = storeToRefs(fileStore);

// WS10: the bare-key view/action shortcuts (1 / 2 / 3 / n / u / e / r) were
// removed. Those actions now live in the command palette (⌘K) and the ⋯ menu.
// The listing keeps only the Finder-style navigation keys, handled directly in
// `keyEvent` below (arrows, Home/End, type-ahead, Enter, ⌘A, Esc).

const route = useRoute();
onBeforeRouteUpdate(() => {
  hideContextMenu();
});

const { t } = useI18n();

const listing = ref<HTMLElement | null>(null);
const scrollSection = ref<HTMLElement | null>(null);

// S7-2: touch-only camera-roll / photo-library upload. `isTouchDevice`
// gates the header affordance; `photoInput` is the hidden <input> it
// triggers (change → existing uploadInput pipeline).
const isTouchDevice = useTouchDevice();
const photoInput = ref<HTMLInputElement | null>(null);

// Retract the mobile details sheet whenever the selection empties — covers
// deselect-all AND navigation (the route watcher clears `selected`), so a
// stale sheet never lingers over the next folder.
watch(
  () => fileStore.selectedCount,
  (n) => {
    if (n === 0) layoutStore.mobileDetailsOpen = false;
  }
);
// v1.3 S6-1: RecycleScroller instance for the virtualized list view.
// Held so we can scroll a previously-opened item back into view when the
// user navigates back from a preview (the recycler's analog of the
// non-virtual `revealPreviousItem` querySelector + scrollIntoView).
const listScroller = ref<{ scrollToItem: (index: number) => void } | null>(
  null
);

// ── Remember scroll position across navigations (per-user pref) ─────────
// The scrolling element differs by view: the list view uses RecycleScroller's
// own scroller; grid/gallery scroll `scrollSection`. Capture/restore both —
// writing scrollTop to the non-scrolling one is a harmless no-op.
const scrollEls = (): HTMLElement[] => {
  const els: HTMLElement[] = [];
  if (scrollSection.value) els.push(scrollSection.value);
  const rec = (listScroller.value as { $el?: HTMLElement } | null)?.$el;
  if (rec && rec !== scrollSection.value) els.push(rec);
  return els;
};

// Per-user "remember position when navigating back" preference (default on).
// When enabled, returning to a folder you were recently in — even across a
// sideways jump to an unrelated directory — restores its scroll position, and a
// same-folder reload (e.g. after an upload finishes) keeps your place instead
// of snapping to the top. A per-folder memory (not a single slot) so either
// side of an A → B → A round trip restores. We record the leaving folder's
// position pre-render (rows still mounted) and restore it on arrival (the `req`
// watcher below).
const rememberParentScroll = computed<boolean>(() =>
  prefs.get<boolean>("nav.rememberParentScroll", true)
);
const folderScroll = useFolderScrollMemory(() => rememberParentScroll.value);

watch(
  () => layoutStore.loading,
  (loading) => {
    // A listing fetch is starting → snapshot the folder we're leaving while its
    // rows are still mounted (so scrollTop is valid). This also captures the
    // position before a same-folder reload. Restoration happens on arrival.
    if (!loading) return;
    const top = Math.max(0, ...scrollEls().map((e) => e.scrollTop));
    folderScroll.record(fileStore.req?.path ?? null, top);
  }
);

// ── Grid/gallery virtualization (CH-1) ──────────────────────────────
// Window the folders + files grids so a huge folder only mounts the tiles
// near the viewport (+ a buffer) instead of the old ever-growing
// `showLimit` slice. `dirsSectionRef`/`filesSectionRef` point at the two
// section WRAPPERS (`.listing-section-pad`) — not the inner grid — so their
// top edge is the section's stable row-0 origin (the wrapper spans the top
// spacer + grid + bottom spacer, total height held constant). The composable
// measures column count + tile height from the grid inside, computes the
// window, and the spacer pads keep the scroll height/position stable. Using
// the wrapper (not the grid, whose top shifts with the top pad) is essential:
// it keeps the window math + lasso hit-test anchored to a fixed origin.
const dirsSectionRef = ref<HTMLElement | null>(null);
const filesSectionRef = ref<HTMLElement | null>(null);
const listingGrid = useListingGrid({
  enabled: () => viewMode.value !== "list",
  gallery: () => viewMode.value === "mosaic gallery",
  scrollEl: () => scrollSection.value,
  dirsSectionEl: () => dirsSectionRef.value,
  filesSectionEl: () => filesSectionRef.value,
  dirsCount: () => items.value.dirs.length,
  filesCount: () => items.value.files.length,
});
const { dirsWin, filesWin } = listingGrid;
const scheduleGridUpdate = throttle(() => listingGrid.update(), 60);

// ── Finder-style keyboard navigation (arrows / type-ahead / Enter). Wired into
//    `keyEvent` below. WS10 removed Quick Look (Space) entirely. ───────────────

// Full visual order (folders then files). Mirrors what the user sees; each item
// carries its `req.items` index, which is what `selected`/`activeIndex` store.
const orderedItems = computed<ResourceItem[]>(() => [
  ...items.value.dirs,
  ...items.value.files,
]);

// Scroll the row/tile with the given req-index into view across all three
// virtualized views (list recycler / windowed grid + gallery).
const scrollActiveIntoView = (reqIndex: number) => {
  void nextTick(() => {
    const sc = scrollSection.value;
    if (!sc) return;
    const sel = `[data-index="${reqIndex}"]`;
    const found = sc.querySelector(sel) as HTMLElement | null;
    if (found) {
      found.scrollIntoView({ block: "nearest" });
      return;
    }
    if (viewMode.value === "list") {
      const pos = listRows.value.findIndex((r) => r.item?.index === reqIndex);
      if (pos >= 0) listScroller.value?.scrollToItem(pos);
      return;
    }
    // Grid/gallery target windowed-out of the DOM: jump proportionally, then
    // settle precisely once the freshly-rendered tile exists.
    const order = orderedItems.value;
    const vpos = order.findIndex((it) => it.index === reqIndex);
    if (vpos < 0) return;
    const frac = order.length > 1 ? vpos / (order.length - 1) : 0;
    sc.scrollTop = frac * Math.max(0, sc.scrollHeight - sc.clientHeight);
    requestAnimationFrame(() => {
      (sc.querySelector(sel) as HTMLElement | null)?.scrollIntoView({
        block: "nearest",
      });
    });
  });
};

const listingNav = useListingNavigation({
  ordered: () => orderedItems.value,
  columnsFor: (it) =>
    it.isDir ? listingGrid.dirsCols.value : listingGrid.filesCols.value,
  grid: () => viewMode.value !== "list",
  reveal: scrollActiveIntoView,
});

// Enter → open the focused item (folder navigates in, file opens its preview).
// Falls back to the first selected item so Enter works after a plain click too.
const openActiveItem = () => {
  let idx = fileStore.activeIndex;
  if (idx < 0 && fileStore.selected.length > 0) idx = fileStore.selected[0];
  const it = idx >= 0 ? fileStore.req?.items[idx] : null;
  if (it) void router.push({ path: it.url });
};

// Type-ahead select: printable keys jump the selection to a matching name.
// Tapping one letter repeatedly cycles through same-prefixed siblings; mixing
// letters builds a prefix that refines in place; the session resets after a
// short idle. All the matching/cycle logic lives in (and is unit-tested via)
// utils/typeahead — a timestamp-guarded session, so a missed timer can never
// silently reset the prefix mid-typing.
const typeahead = new TypeaheadSession();
const typeaheadPush = (ch: string) => {
  const matchIndex = typeahead.push(
    ch,
    orderedItems.value,
    fileStore.activeIndex
  );
  if (matchIndex < 0) return;
  fileStore.selected = [matchIndex];
  fileStore.activeIndex = matchIndex;
  fileStore.anchorIndex = matchIndex;
  scrollActiveIntoView(matchIndex);
};

// ── S4-3 drag-select lasso (grid + gallery only) ───────────────────
// Rubber-band selection bound to #listing's @mousedown. `enabled`
// gates on viewMode (declared later — the closure only reads it at
// mouse-down time, well after init). Writes straight to
// fileStore.selected so the marquee live-updates the listing.
// `hitTest` (CH-1) computes intersections by geometry so tiles that have
// been windowed out of the DOM are still selectable across a long drag.
const dragSelect = useDragSelect({
  enabled: () => viewMode.value !== "list",
  getScrollContainer: () => scrollSection.value,
  getSelection: () => [...fileStore.selected],
  setSelection: (indices) => {
    fileStore.selected = indices;
  },
  hitTest: (rect) =>
    listingGrid.hitTest(
      rect,
      items.value.dirs.map((d) => d.index),
      items.value.files.map((f) => f.index)
    ),
});
onBeforeUnmount(() => dragSelect.cleanup());

const nameSorted = computed(() =>
  fileStore.req ? fileStore.req.sorting.by === "name" : false
);

const sizeSorted = computed(() =>
  fileStore.req ? fileStore.req.sorting.by === "size" : false
);

const modifiedSorted = computed(() =>
  fileStore.req ? fileStore.req.sorting.by === "modified" : false
);

const ascOrdered = computed(() =>
  fileStore.req ? fileStore.req.sorting.asc : false
);

// Custom label for the files root ("My files"), set via the sidebar's
// right-click rename. Falls back to the default at the root.
const { rootLabel } = useRootLabel();

const folderTitle = computed(() => {
  if (!fileStore.req) return "";
  return fileStore.req.name || rootLabel.value || t("sidebar.myFiles");
});

// ── Section title as parent-folder drop + spring-load target (F2) ──
// During a drag, the section-title area (the row that shows the
// current folder name + meta) acts as a shortcut to the PARENT folder:
//   • Drop on it           → move/copy the selection up one level
//   • Hover 2 s during drag → navigate up one level (no drop required)
// Both are gated by the existence of a parent — at the storage root
// we suppress the drop target entirely (no parent to navigate to).
const PARENT_SPRING_MS = 2000;
const sectionDropActive = ref<boolean>(false);
let sectionSpringTimer: number | null = null;
let sectionDragDepth = 0;

const { performDrop: performParentDrop } = useDropTarget();

/** Parent folder URL relative to the current route, or null at root. */
const parentFolderUrl = computed<string | null>(() => {
  if (!fileStore.req?.isDir) return null;
  const here = fileStore.req.url; // ends with "/"
  // Strip trailing slash, then drop the last segment.
  const trimmed = here.endsWith("/") ? here.slice(0, -1) : here;
  const parent = url.removeLastDir(trimmed) + "/";
  // If removing the last segment lands us back at the same place, we
  // were already at the root — nothing to navigate to.
  if (parent === here) return null;
  return parent;
});

/** Click handler for the inline ↑ button. Same destination as the
 *  spring-load drag behavior so users get one mental model regardless
 *  of which input modality they're using. */
const goToParentFolder = () => {
  if (parentFolderUrl.value) router.push({ path: parentFolderUrl.value });
};

// ── Current-folder favorites (v1.3 S3-2) ────────────────────────────
// Star toggle in the section-title eyebrow. Pinning the current
// folder adds it to the sidebar's Favorites list for one-click
// return. Folders only — file pinning belongs in Recents.
const favoritesComposable = useFavorites();
/** The path the favorites composable persists. Uses the listing's
 *  url field (the /files/...-prefixed form) since that's the same
 *  string the ListingItem star uses for child folders. */
const currentFolderPath = computed<string>(() => fileStore.req?.url ?? "");
const currentFolderFavorited = computed<boolean>(() =>
  currentFolderPath.value
    ? favoritesComposable.isFavorited(currentFolderPath.value)
    : false
);
const onCurrentFolderFavToggle = () => {
  if (!currentFolderPath.value) return;
  favoritesComposable.toggle(currentFolderPath.value);
};

const cancelSectionSpring = () => {
  if (sectionSpringTimer !== null) {
    window.clearTimeout(sectionSpringTimer);
    sectionSpringTimer = null;
  }
};

const onSectionDragEnter = (event: DragEvent) => {
  // Gate on the active DRAG set (not the current selection) so a cross-pane
  // drag — whose items live in `draggedItems`, not pane A's `selected` — also
  // arms the parent spring-load when held over the split header (#18).
  if (fileStore.draggedItems.length === 0) return;
  if (!parentFolderUrl.value) return;
  event.preventDefault();
  sectionDragDepth++;
  if (sectionDragDepth === 1) {
    sectionDropActive.value = true;
    // Spring-load: hover for PARENT_SPRING_MS → navigate up.
    sectionSpringTimer = window.setTimeout(() => {
      sectionSpringTimer = null;
      sectionDropActive.value = false;
      sectionDragDepth = 0;
      if (parentFolderUrl.value) router.push({ path: parentFolderUrl.value });
    }, PARENT_SPRING_MS);
  }
};

const onSectionDragOver = (event: DragEvent) => {
  if (fileStore.draggedItems.length === 0) return;
  if (!parentFolderUrl.value) return;
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect =
      event.ctrlKey || event.metaKey ? "copy" : "move";
  }
};

const onSectionDragLeave = () => {
  if (!parentFolderUrl.value) return;
  sectionDragDepth = Math.max(0, sectionDragDepth - 1);
  if (sectionDragDepth === 0) {
    sectionDropActive.value = false;
    cancelSectionSpring();
  }
};

const onSectionDrop = (event: DragEvent) => {
  // Drop wins over spring-load: kill the timer before any conflict
  // prompts so we don't navigate mid-resolve.
  cancelSectionSpring();
  sectionDragDepth = 0;
  sectionDropActive.value = false;
  if (!parentFolderUrl.value) return;
  void performParentDrop(event, parentFolderUrl.value);
};

// ── Pane A cross-pane drop overlay (#17) ─────────────────────────────
// Mirrors ComparePane's `.compare-body--drop`: while an internal selection is
// dragged over pane A's body, draw the same dashed accent frame so pane A reads
// as a drop target too (previously only pane B lit up). Split-only — in single
// pane these handlers early-return, so that path is byte-for-byte unchanged.
// `fileStore.draggedItems` is the shared cross-pane drag set (written by either
// pane's dragstart), so this fires for both A→A and B→A drags.
const paneADropDepth = ref<number>(0);
const paneADropActive = computed(
  () =>
    splitActive.value &&
    paneADropDepth.value > 0 &&
    fileStore.draggedItems.length > 0
);
const onPaneADragEnter = (event: DragEvent) => {
  if (!splitActive.value || fileStore.draggedItems.length === 0) return;
  event.preventDefault();
  paneADropDepth.value++;
};
const onPaneADragOver = (event: DragEvent) => {
  if (!splitActive.value || fileStore.draggedItems.length === 0) return;
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect =
      event.ctrlKey || event.metaKey ? "copy" : "move";
  }
};
const onPaneADragLeave = () => {
  if (paneADropDepth.value > 0) paneADropDepth.value--;
};
const onPaneADrop = (event: DragEvent) => {
  paneADropDepth.value = 0;
  // Only internal cross-pane drags land here; OS-file drops (no draggedItems)
  // are owned by the global document `drop` handler — don't double-handle them.
  if (!splitActive.value || fileStore.draggedItems.length === 0) return;
  // A row (folder into-zone, or a file row's dropAlongside) already handled it.
  if ((event.target as HTMLElement | null)?.closest(".item")) return;
  if (!currentFolderUrl.value) return;
  void performParentDrop(event, currentFolderUrl.value);
};

// Drag-cancel safety net (review #1): Esc-cancelling a drag fires `dragend` but
// NOT `dragleave`/`drop`, so a pending section spring-load timer would otherwise
// still navigate ~PARENT_SPRING_MS later, and the pane-A drop overlay would stay
// lit. Wired to the document `dragend` alongside `resetOpacity` (same rationale).
const onListingDragEnd = () => {
  cancelSectionSpring();
  sectionDropActive.value = false;
  sectionDragDepth = 0;
  paneADropDepth.value = 0;
};

// `currentFolderUrl` — the current folder's url (trailing "/" like the
// ListingItem rows, so a conflict prompt's `to` matches the breadcrumb).
// Destination for an "alongside" drop (below) and for the touch drop path.
const currentFolderUrl = computed<string>(() => fileStore.req?.url ?? "");

// Alongside drop on any row → move into the CURRENT folder via the shared
// useDropTarget.performDrop (target-agnostic; it just takes a destination
// URL). A release on a row's non-into-zone area, or on any file row, routes
// here. No-op if we somehow have no current folder.
const onItemDropAlongside = (event: DragEvent) => {
  if (!currentFolderUrl.value) return;
  void performParentDrop(event, currentFolderUrl.value);
};

// ── Touch drag-and-drop (lifted from ListingItem, CH-2) ─────────────
// HTML5 DnD never fires on touch, so a single pointer-based gesture
// drives the SAME move pipeline (`useDropTarget.performDrop`) — conflict
// prompt, self-drop guard, transfer indicator. This used to be one
// useTouchDrag instance PER row; hoisting it here means the listing owns
// exactly one regardless of how many rows are mounted. Rows forward their
// pointerdown via the `rowPointerDown` event (mouse pointers are ignored
// inside the composable, so desktop is untouched). Drop targets are any
// element carrying `data-drop-url` — folder rows, breadcrumb segments,
// and the current-folder area.
let touchHighlightEl: HTMLElement | null = null;
let touchSpringUrl: string | null = null;
let touchSpringTimer: number | null = null;
// Destination the in-flight touch drop will resolve to, cached on every
// onMove so onDrop matches EXACTLY what was highlighted (the same robustness as
// the desktop cached-into-zone fix). A folder url when over a folder's tight
// into-zone (or a dedicated breadcrumb / current-folder target); the current
// folder ("alongside") when over a folder row OUTSIDE its into-zone; null when
// over nothing droppable (→ no-op drop).
let touchDropUrl: string | null = null;

const clearTouchHighlight = () => {
  if (touchHighlightEl) {
    touchHighlightEl.style.outline = "";
    touchHighlightEl.style.outlineOffset = "";
    touchHighlightEl = null;
  }
};
const cancelTouchSpring = () => {
  if (touchSpringTimer !== null) {
    window.clearTimeout(touchSpringTimer);
    touchSpringTimer = null;
  }
  touchSpringUrl = null;
};
const resolveDropEl = (el: Element | null): HTMLElement | null =>
  (el?.closest?.("[data-drop-url]") as HTMLElement | null) ?? null;

const listingTouchDrag = useTouchDrag<{ index: number }>({
  // The ghost is created before onStart runs (so draggedItems isn't
  // populated yet for this gesture); read the pressed row's name straight
  // off the listing for the single-item label, matching the old per-row
  // behavior. Multi-select uses the snapshot count once it exists.
  ghostLabel: (p) => {
    const c = fileStore.draggedItems.length;
    if (c > 1) return `${c} items`;
    return fileStore.req?.items[p.index]?.name ?? "";
  },
  // Edge auto-scroll during a touch drag targets the ACTUAL scroll
  // container (the recycler in list view, the <section> in grid/gallery) —
  // resolved by ptrScrollEl — rather than #listing, which isn't the scroller.
  scrollEl: () => ptrScrollEl.value,
  onStart: (p) => fileStore.snapshotDragSelection(p.index),
  onMove: (_p, x, y, el) => {
    const raw = resolveDropEl(el);
    const isFolderRow =
      !!raw &&
      raw.classList.contains("item") &&
      raw.getAttribute("data-dir") === "true";
    // Parity with desktop: a folder ROW is an into-folder target ONLY when the
    // finger is in its tight icon+name into-zone. Outside that the drop goes
    // "alongside" into the current folder, so don't highlight/spring the folder.
    // Dedicated drop targets (breadcrumb segments, the current-folder area)
    // aren't `.item` rows and keep their whole-element target.
    const inIntoZone = isFolderRow && resolveRowDropMode(raw!, x, y) === "into";
    const highlightEl = isFolderRow ? (inIntoZone ? raw : null) : raw;

    if (highlightEl !== touchHighlightEl) {
      clearTouchHighlight();
      if (highlightEl) {
        highlightEl.style.outline = "2px solid var(--color-accent, #6e72d9)";
        highlightEl.style.outlineOffset = "-2px";
        touchHighlightEl = highlightEl;
      }
    }

    // Cache where this drop resolves: the highlighted target's url, else the
    // current folder for a folder row we're not "into" ("alongside"), else null.
    if (highlightEl) {
      touchDropUrl = highlightEl.dataset.dropUrl ?? null;
    } else if (isFolderRow) {
      touchDropUrl = currentFolderUrl.value || null;
    } else {
      touchDropUrl = null;
    }

    // Spring-load: hovering a folder row's into-zone (not the current folder)
    // for 2s drills into it so nested drops are possible (F6 parity).
    const springUrl = inIntoZone ? (raw!.dataset.dropUrl ?? null) : null;
    if (springUrl && springUrl !== fileStore.req?.url) {
      if (touchSpringUrl !== springUrl) {
        cancelTouchSpring();
        touchSpringUrl = springUrl;
        touchSpringTimer = window.setTimeout(() => {
          touchSpringTimer = null;
          void router.push({ path: springUrl });
        }, 2000);
      }
    } else {
      cancelTouchSpring();
    }
  },
  onDrop: () => {
    cancelTouchSpring();
    clearTouchHighlight();
    // Use the destination cached during onMove so the drop lands exactly where
    // it was highlighted (no fresh recompute). null = released over nothing
    // droppable → no-op.
    const dest = touchDropUrl;
    touchDropUrl = null;
    if (!dest) return;
    // Touch has no Ctrl/Cmd → always a move. performParentDrop reads the
    // snapshot from fileStore.draggedItems and applies all the usual guards
    // (incl. the from===to short-circuit, so an "alongside" drop is a no-op).
    const synthetic = {
      preventDefault: () => {},
      ctrlKey: false,
      metaKey: false,
    } as unknown as DragEvent;
    void performParentDrop(synthetic, dest);
  },
  onEnd: () => {
    cancelTouchSpring();
    clearTouchHighlight();
    touchDropUrl = null;
    fileStore.draggedItems = [];
    // Swallow the synthetic click the browser fires on the drop row.
    fileStore.suppressClicksUntil = Date.now() + 350;
  },
});

// Forwarded from each row's pointerdown (after the row's own read-only +
// interactive-child guard). Mouse pointers are ignored inside the gesture.
const onItemPointerDown = (event: PointerEvent, index: number) => {
  listingTouchDrag.onPointerDown(event, { index });
};

// ── Rename the currently-viewed folder ─────────────────────────────────
// Surfaced as a "Rename folder" action in the ⋯ menu (header section title
// More dropdown). UX matches the inline row rename in ListingItem: swap
// the h1 for an input, Enter commits, Esc/blur cancels. We can't rename
// at the storage root (no parent to move into), so the action hides via
// `canRenameCurrentFolder` when the folder has no usable parent.
const router = useRouter();
const isRenamingCurrentFolder = ref<boolean>(false);
const folderRenameValue = ref<string>("");
const folderRenameInputEl = ref<HTMLInputElement | null>(null);
let folderRenameSubmitting = false;

// Mirror the current-folder rename into the shared store flag so a background
// listing refresh (e.g. a transfer's incremental reload) defers instead of
// interrupting the rename mid-edit (see Files.vue's reload gate). Row rename /
// new-folder / new-file already surface via layoutStore.currentPromptName.
watch(isRenamingCurrentFolder, (active) => {
  fileStore.inlineEditing = active;
});

const canRenameCurrentFolder = computed<boolean>(() => {
  if (!authStore.user?.perm.rename) return false;
  const req = fileStore.req;
  if (!req || !req.isDir) return false;
  // Don't expose at the storage root — there's no parent to move into.
  // Root URL looks like "/files/" with no folder name.
  if (!req.name) return false;
  return true;
});

const startFolderRename = async () => {
  if (!canRenameCurrentFolder.value || !fileStore.req) return;
  folderRenameValue.value = fileStore.req.name;
  folderRenameSubmitting = false;
  isRenamingCurrentFolder.value = true;
  await nextTick();
  const el = folderRenameInputEl.value;
  if (!el) return;
  el.focus();
  el.select();
};

const cancelFolderRename = () => {
  if (folderRenameSubmitting) return;
  isRenamingCurrentFolder.value = false;
};

const onFolderRenameBlur = () => {
  if (folderRenameSubmitting) return;
  // Small delay so an Enter keydown can commit before the blur cancel
  // fires (input loses focus when the keydown fires too).
  setTimeout(() => {
    if (!folderRenameSubmitting && isRenamingCurrentFolder.value) {
      cancelFolderRename();
    }
  }, 120);
};

const submitFolderRename = async () => {
  if (folderRenameSubmitting || !fileStore.req) return;
  const next = folderRenameValue.value.trim();
  if (next === "" || next === fileStore.req.name) {
    cancelFolderRename();
    return;
  }
  folderRenameSubmitting = true;
  const oldUrl = fileStore.req.url;
  // Strip trailing slash before computing the parent (removeLastDir
  // would otherwise drop the folder name itself, not its parent).
  const trimmed = oldUrl.endsWith("/") ? oldUrl.slice(0, -1) : oldUrl;
  const newUrl = url.removeLastDir(trimmed) + "/" + encodeURIComponent(next);
  try {
    await api.move([{ from: oldUrl, to: newUrl }]);
    // Keep Favorites pointing at this folder (or descendants) pinned —
    // follow the rename so the sidebar link doesn't break.
    favoritesComposable.renamePath(oldUrl, newUrl);
    // The route URL still points at the old name — navigate to the new
    // path so the listing reloads against the renamed folder.
    router.push({ path: newUrl });
    isRenamingCurrentFolder.value = false;
  } catch (e) {
    if (e instanceof Error) $showError(e);
    folderRenameSubmitting = false;
  }
};

const folderMeta = computed(() => {
  const req = fileStore.req;
  if (!req) return "";
  const parts: string[] = [];
  const totalItems = (req.numDirs ?? 0) + (req.numFiles ?? 0);
  parts.push(`${totalItems} ${totalItems === 1 ? "item" : "items"}`);
  // Deliberately NOT showing a folder size: the filesystem only gives us the
  // directory's own inode size (a handful of bytes), not the recursive total
  // of its contents — which was misleading. Computing the real total means a
  // du-style walk of the whole subtree on every open, which isn't worth the
  // latency. Per-file sizes still show on the rows.
  if (req.modified) {
    parts.push(`last updated ${timeAgo(req.modified)}`);
  }
  return parts.join(" · ");
});

// Short item count for the hero crumb row (the full "N items · last updated…"
// string is folderMeta, shown as the count's tooltip).
const folderCountLabel = computed(() => {
  const req = fileStore.req;
  if (!req) return "";
  const total = (req.numDirs ?? 0) + (req.numFiles ?? 0);
  return `${total} ${total === 1 ? "item" : "items"}`;
});

// CH-1: render only the windowed slice of each grid section. Before the
// first measure (window still {0,0}) fall back to a small bootstrap slice
// so there ARE tiles to measure + fill the first paint.
const BOOTSTRAP_TILES = 60;
const dirs = computed(() => {
  const all = items.value.dirs;
  if (dirsWin.end <= dirsWin.start) {
    return all.slice(0, Math.min(all.length, BOOTSTRAP_TILES));
  }
  return all.slice(dirsWin.start, dirsWin.end);
});

const items = computed(() => {
  const dirs: ResourceItem[] = [];
  const files: ResourceItem[] = [];

  fileStore.req?.items.forEach((item) => {
    if (item.isDir) {
      dirs.push(item);
    } else if (
      typeFilter.value === "all" ||
      typeFilterCategory(item) === typeFilter.value
    ) {
      // v2.7 type chips: files are filtered client-side; folders are exempt
      // (hiding the way forward would turn the filter into a trap).
      files.push(item);
    }
  });

  // 2.4.x: the listing's DISPLAY order is decided client-side (primary +
  // optional secondary tiebreaker) so a sort change re-orders instantly,
  // rather than waiting on a server round-trip + silent reload. dirs/files
  // sort independently so the dirs-first grouping isn't disturbed; `req.items`
  // itself is untouched, so each item keeps its server-assigned `index`.
  const primary = {
    by: (fileStore.req?.sorting.by ?? "name") as SortKey,
    asc: fileStore.req?.sorting.asc ?? false,
  };
  const sec = secondarySort.value;
  // Folders sort by their REAL recursive size (the value shown in the Size
  // column), which lives in the lazily-populated folder-size cache — not the
  // listing's inode `size`. Reading `folderSizes.cached` here (via the sort)
  // makes this computed reactive: as each folder's size resolves, the list
  // re-sorts into place. Sizes are prefetched below when size-sort is active.
  const folderSizeOf = (item: ResourceItem): number | undefined =>
    item.isDir
      ? folderSizes.cached(item.url, String(item.modified ?? ""))
      : undefined;
  return {
    dirs: sortListing(dirs, primary, sec, folderSizeOf),
    files: sortListing(files, primary, sec, folderSizeOf),
  };
});

// When sorting by size, prefetch EVERY folder's recursive size (not just the
// visible rows the Size column lazy-loads) so the order is complete and
// correct rather than settling only as the user scrolls. ensureMany skips
// already-cached/in-flight folders and bounds concurrency, and the server
// caches + singleflights, so repeats are cheap. Fires only while size-sort is
// active; rows re-sort reactively as each resolves.
watch(
  () => ({
    by: fileStore.req?.sorting.by,
    dirs: items.value.dirs,
  }),
  ({ by, dirs }) => {
    if (by !== "size") return;
    void folderSizes.ensureMany(
      dirs.map((d) => ({ path: d.url, mod: String(d.modified ?? "") }))
    );
  },
  { immediate: true }
);

const files = computed((): ResourceItem[] => {
  const all = items.value.files;
  if (filesWin.end <= filesWin.start) {
    return all.slice(0, Math.min(all.length, BOOTSTRAP_TILES));
  }
  return all.slice(filesWin.start, filesWin.end);
});

// ── List virtualization (v1.3 S6-1) ─────────────────────────────────
// Only the LIST view virtualizes (uniform 40px rows). Grid + gallery
// tiles wrap to variable heights, so they stay on the incremental
// `showLimit` windowing above — list-first per the locked Stage 6 plan.
const isListVirtual = computed<boolean>(
  () => effectiveViewMode.value === "list"
);

// Flat row model the RecycleScroller consumes: every dir, an optional
// folder/file divider sentinel, then every file. Crucially this is the
// FULL list (no `showLimit` slice) — the recycler only mounts the
// on-screen window, so a 10k-entry array costs almost nothing. Reuses
// `items` (already secondary-sorted + pending-delete-filtered), so sort
// and the delete-undo flow stay correct through virtualization.
// Flat shape uses a nullable `item` (rather than a discriminated union)
// so the template can narrow with a plain `v-else-if="row.item"` — Vue's
// template type-checker narrows truthiness reliably but not always a
// `kind` discriminant across v-else.
interface ListRow {
  /** Stable key for the recycler. */
  id: string;
  /** True for the single folder/file divider sentinel. */
  divider: boolean;
  /** The file/folder for a normal row; null on the divider. */
  item: ResourceItem | null;
  /** Per-row height for the variable-size recycler (item-size=null). Only the
   *  divider sets it (a slim slot); item rows fall back to min-item-size. */
  size?: number;
}

const listRows = computed<ListRow[]>(() => {
  const { dirs, files } = items.value;
  const rows: ListRow[] = [];
  for (const d of dirs) {
    rows.push({ id: "d:" + base64(d.name), divider: false, item: d });
  }
  if (dirs.length > 0 && files.length > 0) {
    // `size` drives the virtual scroller's variable row height (item-size=null):
    // the divider gets a slim slot so folders→files aren't separated by a full
    // 44px row of dead space. Item rows omit `size` → fall back to min-item-size.
    rows.push({ id: "__divider__", divider: true, item: null, size: 18 });
  }
  for (const f of files) {
    rows.push({ id: "f:" + base64(f.name), divider: false, item: f });
  }
  return rows;
});

/** Virtualized analog of `revealPreviousItem`: when the user returns to
 *  a folder with a previously-opened item selected, scroll the recycler
 *  so that row is in view (the off-screen DOM node doesn't exist, so we
 *  can't use scrollIntoView). */
const revealInVirtualList = (): boolean => {
  if (!fileStore.req || !fileStore.oldReq) return false;
  const index = fileStore.selected[0];
  if (index === undefined) return false;
  nextTick(() => {
    const rowIdx = listRows.value.findIndex(
      (r) => r.item !== null && r.item.index === index
    );
    if (rowIdx >= 0) listScroller.value?.scrollToItem(rowIdx);
  });
  return true;
};

const nameIcon = computed(() => {
  if (nameSorted.value && !ascOrdered.value) {
    return "arrow-up";
  }

  return "arrow-down";
});

const sizeIcon = computed(() => {
  if (sizeSorted.value && ascOrdered.value) {
    return "arrow-down";
  }

  return "arrow-up";
});

const modifiedIcon = computed(() => {
  if (modifiedSorted.value && ascOrdered.value) {
    return "arrow-down";
  }

  return "arrow-up";
});

const headerButtons = computed(() => {
  return {
    upload: authStore.user?.perm.create,
    download: authStore.user?.perm.download,
    shell: authStore.user?.perm.execute && enableExec,
    delete: fileStore.selectedCount > 0 && authStore.user?.perm.delete,
    rename: fileStore.selectedCount === 1 && authStore.user?.perm.rename,
    share:
      fileStore.selectedCount === 1 &&
      authStore.user?.perm.share &&
      authStore.user?.perm.download,
    move: fileStore.selectedCount > 0 && authStore.user?.perm.rename,
    copy: fileStore.selectedCount > 0 && authStore.user?.perm.create,
    // Extract (PR #5746) — one OR MORE archives selected (each extracts into
    // its own subfolder), the user can create (extraction writes new files),
    // and the operator hasn't disabled the feature via `--unzipEnabled=false`.
    // A mixed selection (an archive + a non-archive) doesn't qualify: every
    // selected item must be extractable, so the action is unambiguous.
    extract:
      unzipEnabled &&
      !!authStore.user?.perm.create &&
      fileStore.selectedCount >= 1 &&
      fileStore.req != null &&
      fileStore.req.items.length > 0 &&
      fileStore.selected.every((i) =>
        isExtractable(fileStore.req!.items[i]?.name ?? "")
      ),
  };
});

// 1.6.0 batch audio-tag editing — how many of the selected items are taggable
// audio files (MP3 / FLAC). Drives the bulk "Edit tags" entry points (pill +
// context menu). Non-audio items in a mixed selection are simply ignored; the
// editor operates on the audio subset and states the count it'll affect.
const bulkAudioCount = computed(() => {
  const req = fileStore.req;
  if (!req) return 0;
  let n = 0;
  for (const idx of fileStore.selected) {
    const item = req.items[idx];
    if (item && isAudioTaggable(item.name)) n++;
  }
  return n;
});
// Show the bulk action only for a genuine multi-file batch (≥2 audio files);
// a single selected audio file already gets the inline "Edit tags…" path.
const canBulkEditTags = computed(
  () =>
    fileStore.selectedCount > 1 &&
    bulkAudioCount.value >= 2 &&
    !!authStore.user?.perm.modify
);

const isMobile = computed(() => {
  return width.value <= 736;
});

watch(req, () => {
  // New folder data → abandon any in-progress type-ahead prefix.
  typeahead.reset();
  nextTick(() => {
    // Grid/gallery must always re-measure its window for the new data so the
    // correct tiles render — independent of any scroll handling below.
    if (!isListVirtual.value) {
      listingGrid.measure();
      listingGrid.update();
    }

    // A SAME-FOLDER reload (rename / transfer / upload / tag refresh — the data
    // was revalidated in place, the path didn't change) must NOT move the
    // scroll: the virtual scroller keeps the user's scrollTop across the data
    // swap, so they stay exactly where they were. Touching scroll here is what
    // snapped a just-renamed (re-selected) row to the TOP — the disruptive jump
    // (recall() finds nothing for a silent reload, so it fell through to the
    // returning-to-folder reveal, which scrolls selected[0] into view). The
    // restore / reveal below is only meaningful for an actual folder change.
    if (fileStore.oldReq && fileStore.oldReq.path === fileStore.req?.path) {
      return;
    }

    // "Remember position" (pref): returning to a folder we have a remembered
    // scroll position for (a round-trip visit) restores it. Null otherwise.
    const restoreTop = folderScroll.recall(fileStore.req?.path ?? null);
    if (restoreTop !== null) {
      requestAnimationFrame(() => {
        for (const el of scrollEls()) el.scrollTop = restoreTop;
        if (!isListVirtual.value) listingGrid.update();
      });
      return;
    }

    // S6-1: the list view virtualizes — scroll a previously-opened item back
    // into view if any (returning from a preview).
    if (isListVirtual.value) {
      revealInVirtualList();
      return;
    }

    // CH-1: grid/gallery — reveal a previously-opened tile if returning.
    revealPreviousItem();
  });
});

// CH-1: observes the grid/gallery scroll container so InfoPane open/close
// and window resizes re-measure the column count + window. Created on
// mount, disconnected on unmount.
let gridResizeObserver: ResizeObserver | null = null;

onMounted(() => {
  // Dual-pane: load the saved split on/off + pane-B path from the prefs bag.
  panes.restore();

  // Measure the grid columns/tile-height for the first time.
  columnsResize();

  // S6-1: virtualized list skips the window-fill math entirely.
  if (isListVirtual.value) {
    revealInVirtualList();
  } else {
    listingGrid.measure();
    listingGrid.update();
    revealPreviousItem();
  }

  // Add the needed event listeners to the window and document.
  window.addEventListener("keydown", keyEvent);
  document.addEventListener("paste", onPasteUpload);
  window.addEventListener("scroll", scrollEvent);
  window.addEventListener("resize", windowsResize);

  // CH-1: the grid/gallery scroll the <section>, so window-scroll doesn't
  // fire — listen on the section directly to drive the virtualization
  // window, and observe its size so column count tracks width changes
  // (e.g. the InfoPane opening). The list view manages its own recycler.
  const sc = scrollSection.value;
  if (sc) {
    sc.addEventListener("scroll", scheduleGridUpdate, { passive: true });
    if (typeof ResizeObserver !== "undefined") {
      gridResizeObserver = new ResizeObserver(() => listingGrid.update());
      gridResizeObserver.observe(sc);
    }
  }

  if (!authStore.user?.perm.create) return;
  document.addEventListener("dragover", preventDefault);
  document.addEventListener("dragover", onDragScrollOver);
  document.addEventListener("dragenter", dragEnter);
  document.addEventListener("dragleave", dragLeave);
  document.addEventListener("drop", drop);
  document.addEventListener("drop", stopDragScroll);
  document.addEventListener("dragend", stopDragScroll);
  // Safety net for the Copy/Move badge: the source row's dragend normally
  // tears it down, but spring-load navigation can unmount that row mid-drag,
  // so end it at the document level too (idempotent).
  document.addEventListener("drop", endDragBadge);
  document.addEventListener("dragend", endDragBadge);
  // Same safety net for the dragged-items SNAPSHOT (fileStore.draggedItems):
  // the source row's own dragEnd normally clears it, but spring-load navigation
  // can unmount that row mid-drag so its handler never fires — leaking a
  // non-empty snapshot that keeps the listing "busy" and freezes background
  // refreshes. drop/dragend bubble to the document even when the source row is
  // gone, and this runs in the bubble phase AFTER the row/target drop handlers
  // have already captured their own copy of the snapshot, so clearing it here
  // can't empty an in-flight move.
  document.addEventListener("drop", clearDragSnapshot);
  document.addEventListener("dragend", clearDragSnapshot);
  // Clear the drag DIM (rows set to 50% opacity on dragenter) when a drag ends
  // WITHOUT a drop — e.g. the user presses Esc to cancel. Esc fires `dragend`
  // but not `drop`/`dragleave`, so without this the hovered/dragged rows stayed
  // dimmed until the next interaction. (drop already calls resetOpacity itself.)
  document.addEventListener("dragend", resetOpacity);
  // Cancel a pending parent spring-load + clear the pane-A drop overlay on an
  // Esc-cancelled drag (review #1).
  document.addEventListener("dragend", onListingDragEnd);
  // Esc-cancel safety net: some browsers (and source-row unmount races) don't
  // fire `dragend` when a drag is cancelled with Escape, which strands the
  // Copy/Move badge + the drag snapshot (the snapshot freezes background
  // refreshes). Catch Escape at the document in the CAPTURE phase — it fires
  // even while the native drag owns the pointer — and run the same idempotent
  // teardown as the dragend nets above, but only while a drag is in progress.
  document.addEventListener("keydown", onDragCancelKey, true);
});

onBeforeUnmount(() => {
  // Remove event listeners before destroying this page.
  window.removeEventListener("keydown", keyEvent);
  document.removeEventListener("paste", onPasteUpload);
  window.removeEventListener("scroll", scrollEvent);
  window.removeEventListener("resize", windowsResize);

  // Don't leave the shared inline-edit flag stuck on if we unmount mid-rename
  // (e.g. navigation), which would otherwise defer every future refresh.
  fileStore.inlineEditing = false;

  scrollSection.value?.removeEventListener("scroll", scheduleGridUpdate);
  scheduleGridUpdate.cancel();
  gridResizeObserver?.disconnect();
  gridResizeObserver = null;

  if (authStore.user && !authStore.user?.perm.create) return;
  document.removeEventListener("dragover", preventDefault);
  document.removeEventListener("dragover", onDragScrollOver);
  document.removeEventListener("dragenter", dragEnter);
  document.removeEventListener("dragleave", dragLeave);
  document.removeEventListener("drop", drop);
  document.removeEventListener("drop", stopDragScroll);
  document.removeEventListener("dragend", stopDragScroll);
  document.removeEventListener("drop", endDragBadge);
  document.removeEventListener("dragend", endDragBadge);
  document.removeEventListener("drop", clearDragSnapshot);
  document.removeEventListener("dragend", clearDragSnapshot);
  document.removeEventListener("dragend", resetOpacity);
  document.removeEventListener("dragend", onListingDragEnd);
  document.removeEventListener("keydown", onDragCancelKey, true);
  stopDragScroll();
  endDragBadge();
});

const base64 = (name: string) => Base64.encodeURI(name);

// ── Type filter chips (v2.7) ─────────────────────────────────────────
// Session-scoped, resets on navigation: the filter answers "show me the
// photos in THIS mess", it isn't a saved preference.
const typeFilter = ref<TypeFilterKey | "all">("all");

const typeFilterCounts = computed(() => {
  const counts = new Map<TypeFilterKey, number>();
  for (const it of fileStore.req?.items ?? []) {
    const c = typeFilterCategory(it);
    if (c !== null) counts.set(c, (counts.get(c) ?? 0) + 1);
  }
  return counts;
});

const typeChipList = computed(() =>
  TYPE_FILTER_ORDER.filter((k) => typeFilterCounts.value.has(k)).map((k) => ({
    key: k,
    label: TYPE_FILTER_LABELS[k],
    count: typeFilterCounts.value.get(k) ?? 0,
  }))
);

// Chips only earn their row when the folder actually mixes categories; a
// folder of nothing-but-photos has nothing to filter. Hidden in split (no
// room) — the filter itself also resets on any navigation below.
const showTypeChips = computed(
  () => !splitActive.value && typeFilterCounts.value.size >= 2
);

const setTypeFilter = (key: TypeFilterKey | "all") => {
  // Clicking the active chip toggles back to All.
  typeFilter.value = typeFilter.value === key && key !== "all" ? "all" : key;
};

// New folder, clean slate — and drop any selection the filter change just
// hid (invisible selected rows would make ⌘-actions feel haunted).
watch(
  () => fileStore.req?.url,
  () => {
    typeFilter.value = "all";
  }
);
watch(typeFilter, () => {
  fileStore.selected = [];
  fileStore.activeIndex = -1;
  fileStore.anchorIndex = -1;
});

// Quick Look: the peek shows "pane A's single selected item" and follows the
// selection while open (arrows keep working — QuickPeek lets them through).
const quickPeek = useQuickPeek();
const singleSelectedItem = (): ResourceItem | null =>
  fileStore.selected.length === 1
    ? (fileStore.req?.items[fileStore.selected[0]] ?? null)
    : null;

const keyEvent = (event: KeyboardEvent) => {
  // No prompts are shown
  if (layoutStore.currentPrompt !== null) {
    return;
  }

  // Dual-pane (R4): F6 toggles the active pane (the OS convention for cycling
  // panes). Handled here — before the bail below — so it works whichever pane
  // is active; the typing guard keeps it inert while editing a field.
  if (splitActive.value && event.key === "F6") {
    const tgt = event.target as HTMLElement | null;
    const tg = tgt?.tagName?.toLowerCase();
    if (tg !== "input" && tg !== "textarea" && !tgt?.isContentEditable) {
      event.preventDefault();
      panes.setActive(panes.activePane === "a" ? "b" : "a");
      return;
    }
  }

  // Dual-pane (R4): when the split is open and pane B is the active pane, it
  // owns the keyboard — ComparePane has its own handler. Bail here so the two
  // don't both act. When pane A is active this guard is false, so pane A's
  // keyboard behaves exactly as it did before the split existed.
  if (splitActive.value && panes.activePane === "b") {
    return;
  }

  if (event.key === "Escape") {
    // A pending CUT is an armed state (its rows render dimmed) — Esc disarms
    // it. A pending COPY has no visual state and survives, so you can still
    // paste it elsewhere after clearing a selection.
    if (clipboardStore.key === "cut") clipboardStore.resetClipboard();
    // Reset files selection + keyboard cursor.
    fileStore.selected = [];
    fileStore.activeIndex = -1;
    fileStore.anchorIndex = -1;
    typeahead.reset();
  }

  // ── Finder-style keyboard navigation (arrows / Home / End / Enter /
  //    type-ahead). WS10 removed the bare-key Delete/F2/Space actions + the
  //    global dispatcher shortcuts; these navigation keys are the only ones the
  //    listing still owns. Non-modifier keys, handled BEFORE the ctrl/meta
  //    guard below, and skipped while typing in a field. ────────────────────
  {
    const navTarget = event.target as HTMLElement | null;
    const navTag = navTarget?.tagName?.toLowerCase();
    const typing =
      navTag === "input" ||
      navTag === "textarea" ||
      navTarget?.isContentEditable === true;
    const noChord = !event.metaKey && !event.ctrlKey && !event.altKey;
    if (!typing && noChord) {
      switch (event.key) {
        case "ArrowDown":
          event.preventDefault();
          listingNav.move("down", event.shiftKey);
          return;
        case "ArrowUp":
          event.preventDefault();
          listingNav.move("up", event.shiftKey);
          return;
        case "ArrowRight":
          event.preventDefault();
          listingNav.move("right", event.shiftKey);
          return;
        case "ArrowLeft":
          event.preventDefault();
          listingNav.move("left", event.shiftKey);
          return;
        case "Home":
          event.preventDefault();
          listingNav.move("home", event.shiftKey);
          return;
        case "End":
          event.preventDefault();
          listingNav.move("end", event.shiftKey);
          return;
        case "Enter":
          event.preventDefault();
          openActiveItem();
          return;
        case "/":
          // V2 #3: "/" reloads the current folder (was the "R" shortcut). Sits
          // before the type-ahead default so it isn't swallowed as a search key.
          event.preventDefault();
          fileStore.reload = true;
          return;
        case "Delete": {
          // 2.4.0 Stage 2: the Delete (Del / Fn+⌫) key moves the selection to
          // the Trash; holding SHIFT deletes permanently. Both go through the
          // same confirm dialog — only the wording and the commit differ.
          // Backspace is deliberately NOT bound: delete is destructive enough
          // that the user asked to drop the easy-to-hit accidental trigger.
          if (!authStore.user?.perm.delete) return;
          const delItems = collectSelectedDeleteItems();
          if (delItems.length === 0) return;
          event.preventDefault();
          openDeleteConfirm(delItems, event.shiftKey);
          return;
        }
        case " ": {
          // Quick Look (v2.7, deliberate re-add after WS10 dropped the old
          // one): Space with a SINGLE selected item opens the peek overlay.
          // Mid-type-ahead a space still joins the search prefix ("My Doc"),
          // so type-ahead keeps priority. No selection → let the browser
          // scroll as usual. Closing is handled by QuickPeek itself (its
          // capture-phase handler eats Space/Esc while open).
          if (typeahead.isActive()) {
            event.preventDefault();
            typeaheadPush(event.key);
            return;
          }
          const peekItem = singleSelectedItem();
          if (!peekItem) return;
          event.preventDefault();
          quickPeek.toggle(singleSelectedItem);
          return;
        }
        default:
          // Type-ahead — printable single chars only, and not when the global
          // shortcut dispatcher already claimed the key (it ran first and
          // called preventDefault()).
          //
          // V3-C #25: a space joins the prefix ONLY mid-session (a name like
          // "My Doc" needs it), never as the first key — a leading space
          // matches nothing and Space owns no other bare-key action here
          // (WS10 removed Quick Look). `/\S/` still admits every printable
          // non-space char as a session starter.
          if (
            !event.defaultPrevented &&
            event.key.length === 1 &&
            (/\S/.test(event.key) ||
              (event.key === " " && typeahead.isActive()))
          ) {
            event.preventDefault();
            typeaheadPush(event.key);
            return;
          }
      }
    }
  }

  // Ctrl is pressed
  if (!event.ctrlKey && !event.metaKey) {
    return;
  }

  // G2 + general hygiene: don't hijack modifier shortcuts when the user
  // is typing in an input / textarea / contenteditable. Cmd+A, Cmd+C,
  // Cmd+X, Cmd+V all have native meanings inside text fields that we
  // shouldn't clobber. The shortcut dispatcher in `useShortcuts` skips
  // typing targets for non-modifier keys; this is the equivalent for
  // the modifier shortcuts handled directly here.
  const target = event.target as HTMLElement | null;
  const tag = target?.tagName?.toLowerCase();
  if (tag === "input" || tag === "textarea" || target?.isContentEditable) {
    return;
  }

  switch (event.key) {
    case "f":
    case "F":
      if (event.shiftKey) {
        event.preventDefault();
        layoutStore.showHover("search");
      }
      break;
    case "c":
    case "C":
    case "x":
    case "X":
      // ⌘⇧C — copy the selected item's path (the clipboard-copy sibling;
      // same action as the InfoPane/context-menu "Copy path").
      if (event.key.toLowerCase() === "c" && event.shiftKey) {
        const it =
          fileStore.selectedCount === 1
            ? (fileStore.req?.items[fileStore.selected[0]] ?? null)
            : null;
        if (it) {
          event.preventDefault();
          void copyItemPath(it);
        }
        break;
      }
      clipboardCapture(event.key.toLowerCase() === "x" ? "cut" : "copy", event);
      break;
    case "v":
    case "V":
      void paste();
      break;
    case "a":
      event.preventDefault();
      for (const file of items.value.files) {
        if (fileStore.selected.indexOf(file.index) === -1) {
          fileStore.selected.push(file.index);
        }
      }
      for (const dir of items.value.dirs) {
        if (fileStore.selected.indexOf(dir.index) === -1) {
          fileStore.selected.push(dir.index);
        }
      }
      break;
    case "s":
      event.preventDefault();
      document.getElementById("download-button")?.click();
      break;
  }
};

const preventDefault = (event: Event) => {
  // Wrapper around prevent default.
  event.preventDefault();
  // Trail the live Copy/Move badge with the cursor during a listing drag
  // (the badge element itself is created by the source row's dragstart). The
  // drop's copy/move decision reads the same ctrl/meta, so they stay in sync.
  if (fileStore.draggedItems.length > 0) {
    const e = event as DragEvent;
    moveDragBadge(e.ctrlKey || e.metaKey, e.clientX, e.clientY);
  }
};

// Document-level fallback to clear the internal-drag snapshot once a drag ends
// (see the drop/dragend listeners in onMounted). Idempotent.
const clearDragSnapshot = () => {
  if (fileStore.draggedItems.length > 0) fileStore.draggedItems = [];
};

// Esc-cancel safety net (registered as a capture keydown in onMounted): when a
// drag is cancelled with Escape and the browser doesn't fire `dragend` — or the
// source row unmounted mid-drag — the Copy/Move badge + drag snapshot get
// stranded on screen. Re-run the same idempotent teardown the dragend nets do.
// Guarded on an active drag so a plain Escape (clear selection) is unaffected.
const onDragCancelKey = (event: KeyboardEvent) => {
  if (event.key !== "Escape") return;
  if (fileStore.draggedItems.length === 0) return;
  endDragBadge();
  clearDragSnapshot();
  resetOpacity();
  onListingDragEnd();
  stopDragScroll();
};

// HTML5 (mouse) drag edge auto-scroll. Touch drag already auto-scrolls via
// useTouchDrag; this brings the same behavior to mouse dragging so that
// dragging a row (or OS files) toward the top/bottom edge gently scrolls the
// listing — making deep drops reachable without letting go. Targets the real
// scroll container (the recycler in list view, the <section> in grid/gallery)
// via ptrScrollEl, matching the touch path. Speed is intentionally gentler
// than touch ("not too quickly").
const DRAG_EDGE = 56; // px from an edge where auto-scroll engages
const DRAG_EDGE_SPEED = 8; // max px/frame (vs touch's 12)
let dragScrollY = 0; // latest pointer Y during a drag
let dragScrollDir = 0; // -1 = up, +1 = down, 0 = idle
let dragScrollRaf: number | null = null;

const dragScrollTick = () => {
  dragScrollRaf = null;
  const el = ptrScrollEl.value;
  if (!el || dragScrollDir === 0) return;
  const rect = el.getBoundingClientRect();
  // How far the pointer is *into* the edge band → proportional speed ramp,
  // so it eases in near the boundary and tops out at DRAG_EDGE_SPEED.
  // Clamp the "into the band" depth to DRAG_EDGE so the speed never exceeds
  // DRAG_EDGE_SPEED — the pointer can sit *past* the edge (over the header
  // above the list / a bar below it), which would otherwise ramp dy beyond the
  // cap and scroll faster than the intended gentle pace.
  let dy = 0;
  if (dragScrollDir < 0) {
    const into = Math.min(DRAG_EDGE, DRAG_EDGE - (dragScrollY - rect.top));
    if (into > 0) dy = -Math.ceil((into / DRAG_EDGE) * DRAG_EDGE_SPEED);
  } else {
    const into = Math.min(DRAG_EDGE, DRAG_EDGE - (rect.bottom - dragScrollY));
    if (into > 0) dy = Math.ceil((into / DRAG_EDGE) * DRAG_EDGE_SPEED);
  }
  if (dy !== 0) {
    el.scrollTop += dy;
    dragScrollRaf = requestAnimationFrame(dragScrollTick);
  } else {
    dragScrollDir = 0;
  }
};

const onDragScrollOver = (event: DragEvent) => {
  const el = ptrScrollEl.value;
  if (!el) return;
  const rect = el.getBoundingClientRect();
  dragScrollY = event.clientY;
  // Only engage when the pointer is horizontally over the scroll area, so a
  // drag across the sidebar/info-pane doesn't scroll the listing.
  const insideX = event.clientX >= rect.left && event.clientX <= rect.right;
  let dir = 0;
  if (insideX) {
    if (dragScrollY < rect.top + DRAG_EDGE) dir = -1;
    else if (dragScrollY > rect.bottom - DRAG_EDGE) dir = 1;
  }
  dragScrollDir = dir;
  if (dir !== 0 && dragScrollRaf === null) {
    dragScrollRaf = requestAnimationFrame(dragScrollTick);
  }
};

const stopDragScroll = () => {
  dragScrollDir = 0;
  if (dragScrollRaf !== null) {
    cancelAnimationFrame(dragScrollRaf);
    dragScrollRaf = null;
  }
};

/**
 * Capture the current selection into the app clipboard (NOT the OS clipboard —
 * items are pasted via the background transfer pipeline). `mode` is explicit so
 * the context menu can call this without synthesizing keyboard events.
 * Permission gates mirror the backend's transfer checks: a cut pastes as a
 * MOVE (perm.rename), a copy pastes as a COPY (perm.create). With nothing
 * selected (or no permission) this no-ops WITHOUT preventDefault, so ⌘C still
 * performs the browser's native text-selection copy.
 */
const clipboardCapture = (mode: "copy" | "cut", event?: Event): void => {
  if (fileStore.req === null) return;
  if (mode === "cut" && !authStore.user?.perm.rename) return;
  if (mode === "copy" && !authStore.user?.perm.create) return;

  const items = [];
  for (const i of fileStore.selected) {
    items.push({
      from: fileStore.req.items[i].url,
      name: fileStore.req.items[i].name,
      size: fileStore.req.items[i].size,
      modified: fileStore.req.items[i].modified,
    });
  }
  if (items.length === 0) return;

  event?.preventDefault();
  clipboardStore.$patch({
    key: mode,
    items,
    path: route.path,
  });
};

/**
 * Paste the app clipboard into `dest` (a folder URL; defaults to the current
 * folder). Runs through the shared background transfer pipeline, so a
 * same-volume cut→paste lands on the 2.3.0 fast lane automatically.
 *
 * Same-folder handling (Stage 1):
 *   - CUT pasted back into its source folder is a NO-OP that just disarms the
 *     clipboard (Finder semantics) — previously this moved every item onto a
 *     "(1)" suffix of itself, effectively renaming the originals.
 *   - COPY pasted into its source folder duplicates every item with the
 *     backend's "(N)" suffix directly — no conflict prompt. Every item
 *     trivially collides with itself there, and surfacing "Override" for a
 *     self-copy is a destructive trap, so keep-both is the only resolution.
 */
const paste = async (dest?: string) => {
  if (clipboardStore.items.length === 0) return;

  const rawDest = dest ?? route.path;
  const path = rawDest.endsWith("/") ? rawDest : rawDest + "/";
  const clipSrc = clipboardStore.path
    ? clipboardStore.path.endsWith("/")
      ? clipboardStore.path
      : clipboardStore.path + "/"
    : "";
  const samePlace = clipSrc === path;

  const isMove = clipboardStore.key === "cut";
  const kind: "move" | "copy" = isMove ? "move" : "copy";

  if (isMove && samePlace) {
    clipboardStore.resetClipboard();
    return;
  }

  const items: any[] = [];
  for (const item of clipboardStore.items) {
    const from = item.from.endsWith("/") ? item.from.slice(0, -1) : item.from;
    const to = path + encodeURIComponent(item.name);
    items.push({
      from,
      to,
      name: item.name,
      size: item.size,
      modified: item.modified,
      overwrite: false,
      rename: samePlace,
    });
  }

  if (items.length === 0) {
    return;
  }

  // Run the paste through the SHARED background transfer — the same path the
  // move/copy tool and drag-drop use. The floating transfer dock then (a) shows
  // the progress notification the user expects and (b) refreshes the listing
  // when the job settles, so the pasted file becomes visible without a manual
  // reload. (Previously paste called api.move/api.copy directly: no dock, and
  // the refresh hung on a one-off `fileStore.reload` flag.) Per-item overwrite /
  // rename flags set during conflict resolution are carried through by
  // `startTransfer` → `toTransferItems`.
  const run = () => {
    if (items.length === 0) return;
    void startTransfer(kind, items)
      .then(() => {
        // Selecting the pasted items in the destination is handled centrally
        // when the job settles (TransferDock), using the server's resolved
        // destination names — so it works whether you paste in place (the new
        // copies get a "(1)" suffix and the originals drop out) or after
        // navigating to another folder. A cut+paste consumes the clipboard; a
        // copy+paste keeps it so you can paste again elsewhere.
        if (isMove) clipboardStore.resetClipboard();
      })
      .catch($showError);
  };

  // Same-folder copy: every item collides with itself, so skip the conflict
  // prompt and duplicate with the backend "(N)" suffix (rename was set above).
  if (samePlace) {
    run();
    return;
  }

  const conflict = await upload.checkMoveConflict(items, path);

  if (conflict.length > 0) {
    // Paste path: source is the clipboard's origin folder, target is
    // the current route. clipboardStore.path is the directory the cut
    // / copied items came from.
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
        from: clipboardStore.path,
        to: path,
      },
      confirm: (ev: Event, result: Array<ConflictingResource>) => {
        ev.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            items[item.index].rename = true;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            items[item.index].overwrite = true;
          } else {
            // Skipped (this is what "Skip all conflicting files" produces for
            // every row) — drop it from the batch.
            items.splice(item.index, 1);
          }
        }
        if (items.length > 0) {
          run();
        } else {
          // Every conflicting item was skipped, so there's nothing left to
          // transfer. Without this the dialog just closed silently and the
          // user had no idea whether anything happened (the reported bug).
          $showSuccess(
            `All conflicting items were skipped — nothing was ${
              isMove ? "moved" : "copied"
            }.`
          );
        }
      },
    });

    return;
  }

  run();
};

const columnsResize = () => {
  // CH-1: recompute grid columns + tile height + the visible window from
  // the live container width. CSS Grid still sizes the tiles (auto-fill
  // minmax in listing.css); this only refreshes the virtualization window.
  listingGrid.measure();
  listingGrid.update();
};

// Window-level scroll fallback. Grid/gallery scroll the <section> (we bind
// a listener to it directly in onMounted), but keep this so any ancestor
// scroll still refreshes the window. The list view manages its own.
const scrollEvent = throttle(() => {
  if (isListVirtual.value) return;
  listingGrid.update();
}, 100);

const dragEnter = () => {
  dragCounter.value++;

  // When the user starts dragging an item, put every
  // file on the listing with 50% opacity.
  const items = document.getElementsByClassName("item");

  Array.from(items).forEach((file: Element) => {
    // V3-B #4: never re-dim the active drop target. dragenter fires on every
    // child element the cursor crosses, so this runs repeatedly during a drag.
    // ListingItem.enterIntoZone set the hovered folder to opacity:1 and its
    // `inIntoZone` guard won't re-assert — so re-dimming it here is exactly
    // what made the highlighted folder flicker bright→dim. Skipping rows that
    // carry `item--drop-into` lets the highlight and the spring-load ring
    // coexist, while rows virtualized in mid-drag still get dimmed.
    if (file.classList.contains("item--drop-into")) return;
    (file as HTMLElement).style.opacity = "0.5";
  });
};

const dragLeave = () => {
  dragCounter.value--;

  if (dragCounter.value == 0) {
    resetOpacity();
    // The drag fully left the document. An OS-file drag that exits the window
    // without dropping fires neither `drop` nor `dragend` on us, so this is the
    // only signal to halt the edge auto-scroll rAF (otherwise it busy-loops).
    stopDragScroll();
  }
};

const drop = async (event: DragEvent) => {
  event.preventDefault();
  dragCounter.value = 0;
  resetOpacity();

  const dt = event.dataTransfer;
  let el: HTMLElement | null = event.target as HTMLElement;

  if (fileStore.req === null || dt === null || dt.files.length <= 0) return;

  for (let i = 0; i < 5; i++) {
    if (el !== null && !el.classList.contains("item")) {
      el = el.parentElement;
    }
  }

  const files: UploadList = (await upload.scanFiles(dt)) as UploadList;

  // Dual-pane: when the OS-file drop lands inside pane B (ComparePane), the base
  // destination is pane B's folder, not pane A's route path. This single global
  // handler catches every OS-file drop (it's on `document`); without this, a drop
  // on pane B's empty space or a file row fell through to pane A's path. A drop
  // landing ON a folder row still uploads into that row's folder — resolved just
  // below from its `data-drop-url`, which is already pane-correct.
  const inPaneB =
    (event.target as HTMLElement | null)?.closest?.(".compare-pane") != null;
  const basePath = inPaneB ? panes.secondaryPath : route.path;
  let path = basePath.endsWith("/") ? basePath : basePath + "/";

  // Upload INTO a folder ONLY when the cursor is over its icon + name — the same
  // shared `resolveRowDropMode` hit-test that draws the highlight (path #4 of the
  // four drop surfaces; see utils/dropZone). Anywhere else on the row (or empty
  // space) keeps `path` as the current directory, so the file uploads "alongside".
  // The target folder's url is the row's `data-drop-url` (set only for droppable,
  // non-read-only folders) — no Vue-internals poke.
  const intoFolderUrl =
    el !== null &&
    el.classList.contains("item") &&
    resolveRowDropMode(el, event.clientX, event.clientY) === "into"
      ? el.dataset.dropUrl
      : undefined;
  if (intoFolderUrl) {
    path = intoFolderUrl;

    try {
      (await api.fetch(path)).items;
    } catch (error: any) {
      $showError(error);
      return;
    }
  }

  const conflict = await upload.checkConflict(files, path);

  // Build a preselect list of every uploaded file's destination path.
  // For folder-uploads, `fullPath` carries the relative path inside
  // the dropped folder (already decoded from webkitRelativePath); for
  // plain file drops, `name` is the bare filename (also decoded).
  const buildPreselect = (sourceFiles: typeof files) =>
    sourceFiles.map((f) => removePrefix(path) + (f.fullPath || f.name));

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
        isUploadAction: true,
        to: path,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            continue;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            files[item.index].overwrite = true;
          } else {
            files.splice(item.index, 1);
          }
        }
        if (files.length > 0) {
          upload.handleFiles(files, path, true);
          // Re-select against the post-conflict-resolution survivors
          // so skipped files don't end up "selected but missing". Preselect is
          // a pane-A concept; skip it for pane-B uploads (those refresh pane B).
          if (!inPaneB) fileStore.setPreselect(buildPreselect(files));
        }
      },
    });

    return;
  }

  upload.handleFiles(files, path);
  if (!inPaneB) fileStore.setPreselect(buildPreselect(files));
};

// ── Paste-to-upload (v2.7) ───────────────────────────────────────────
// ⌘V with FILES on the OS clipboard (a screenshot, a Finder copy) uploads
// them into the active pane's folder. Registered on `document` (like the OS
// drop handler) so it works wherever focus sits, with the same guards the
// keyboard handler uses. The app's own cut/copy clipboard keeps priority:
// when it's armed, ⌘V means "paste those items" (handled in keyEvent) and
// this handler stays out of the way.
const onPasteUpload = async (event: ClipboardEvent) => {
  if (layoutStore.currentPrompt !== null) return;
  if (!authStore.user?.perm.create) return;
  if (clipboardStore.key !== "") return; // internal clipboard wins
  const t = event.target as HTMLElement | null;
  const tag = t?.tagName?.toLowerCase();
  if (tag === "input" || tag === "textarea" || t?.isContentEditable) return;
  const clipFiles = event.clipboardData?.files;
  if (!clipFiles || clipFiles.length === 0) return; // plain text — not ours
  event.preventDefault();

  // Paste lands in the ACTIVE pane's folder (split) or the current route.
  const inPaneB = splitActive.value && panes.activePane === "b";
  const basePath = inPaneB ? panes.secondaryPath : route.path;
  const path = basePath.endsWith("/") ? basePath : basePath + "/";

  const now = new Date();
  const uploadFiles: UploadList = [];
  for (let i = 0; i < clipFiles.length; i++) {
    const file = clipFiles[i];
    uploadFiles.push({
      file,
      // Generic clipboard names ("image.png") get a timestamp so repeat
      // pastes don't fight the conflict dialog every time.
      name: pastedFileName(file.name, now),
      size: file.size,
      isDir: false,
    });
  }

  const buildPreselect = (sourceFiles: typeof uploadFiles) =>
    sourceFiles.map((f) => removePrefix(path) + f.name);

  const conflict = await upload.checkConflict(uploadFiles, path);
  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
        isUploadAction: true,
        to: path,
      },
      confirm: (e: Event, result: Array<ConflictingResource>) => {
        e.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            continue;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            uploadFiles[item.index].overwrite = true;
          } else {
            uploadFiles.splice(item.index, 1);
          }
        }
        if (uploadFiles.length > 0) {
          upload.handleFiles(uploadFiles, path, true);
          if (!inPaneB) fileStore.setPreselect(buildPreselect(uploadFiles));
        }
      },
    });
    return;
  }

  upload.handleFiles(uploadFiles, path);
  if (!inPaneB) fileStore.setPreselect(buildPreselect(uploadFiles));
};

const uploadInput = async (event: Event) => {
  const files = (event.currentTarget as HTMLInputElement)?.files;
  if (files === null) return;

  const folder_upload = !!files[0].webkitRelativePath;

  const uploadFiles: UploadList = [];
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const fullPath = folder_upload ? file.webkitRelativePath : undefined;
    uploadFiles.push({
      file,
      name: file.name,
      size: file.size,
      isDir: false,
      fullPath,
    });
  }

  const path = route.path.endsWith("/") ? route.path : route.path + "/";
  const conflict = await upload.checkConflict(uploadFiles, path);

  // Mirror dropUpload's preselect behavior so users see their freshly-
  // uploaded files highlighted in the destination — consistent UX
  // regardless of which upload entry point they used.
  const buildPreselect = (sourceFiles: typeof uploadFiles) =>
    sourceFiles.map((f) => removePrefix(path) + (f.fullPath || f.name));

  if (conflict.length > 0) {
    layoutStore.showHover({
      prompt: "resolve-conflict",
      props: {
        conflict: conflict,
        isUploadAction: true,
        to: path,
      },
      confirm: (event: Event, result: Array<ConflictingResource>) => {
        event.preventDefault();
        layoutStore.closeHovers();
        for (let i = result.length - 1; i >= 0; i--) {
          const item = result[i];
          if (item.checked.length == 2) {
            continue;
          } else if (item.checked.length == 1 && item.checked[0] == "origin") {
            uploadFiles[item.index].overwrite = true;
          } else {
            uploadFiles.splice(item.index, 1);
          }
        }
        if (uploadFiles.length > 0) {
          upload.handleFiles(uploadFiles, path, true);
          fileStore.setPreselect(buildPreselect(uploadFiles));
        }
      },
    });

    return;
  }

  upload.handleFiles(uploadFiles, path);
  fileStore.setPreselect(buildPreselect(uploadFiles));
};

const resetOpacity = () => {
  const items = document.getElementsByClassName("item");

  Array.from(items).forEach((file: Element) => {
    (file as HTMLElement).style.opacity = "1";
    // Clear any lingering drop-into highlight (covers Esc-cancel while the
    // cursor was still inside a folder's into-zone).
    file.classList.remove("item--drop-into");
  });
};

const sort = (by: string) => {
  let asc = false;

  if (by === "name") {
    if (nameIcon.value === "arrow-up") {
      asc = true;
    }
  } else if (by === "size") {
    if (sizeIcon.value === "arrow-up") {
      asc = true;
    }
  } else if (by === "modified") {
    if (modifiedIcon.value === "arrow-up") {
      asc = true;
    }
  } else if (by === "extension") {
    // v1.3 S3-5: default to ascending alphabetical extension order
    // when first selected; subsequent clicks toggle direction. No
    // column-header for extension yet (S3-4 popover will add proper
    // direction toggling for it), so the cycle-button entry point
    // just commits ascending the first time.
    asc = true;
  }

  // Delegate to the shared dispatcher (optimistic re-sort + persist). The
  // asc above is computed from the CURRENT sort icons, so re-clicking a
  // column toggles its direction.
  void sortRaw(by as SortKey, asc);
};

const windowsResize = throttle(() => {
  columnsResize();
  width.value = window.innerWidth;

  // Listing element is not displayed
  if (listing.value == null) return;

  // How much every listing item affects the window height
  setItemWeight();

  // Fill but not fit the window
  fillWindow();
}, 100);

const download = () => {
  if (fileStore.req === null) return;

  if (
    fileStore.selectedCount === 1 &&
    !fileStore.req.items[fileStore.selected[0]].isDir
  ) {
    api.download(null, fileStore.req.items[fileStore.selected[0]].url);
    return;
  }

  // Nothing selected → confirm before downloading the entire folder
  if (fileStore.selectedCount === 0) {
    const folderName = fileStore.req.name || "this folder";
    const ok = window.confirm(
      `Nothing is selected. This will download the entire contents of "${folderName}" as an archive. Continue?`
    );
    if (!ok) return;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: any) => {
      layoutStore.closeHovers();

      const files = [];

      if (fileStore.selectedCount > 0 && fileStore.req !== null) {
        for (const i of fileStore.selected) {
          files.push(fileStore.req.items[i].url);
        }
      } else {
        files.push(route.path);
      }

      api.download(format, ...files);
    },
  });
};

// ── Per-folder view-mode memory (v1.3 S3-3) ─────────────────────────
// localStorage map keyed by folder URL. Per-folder overrides win over
// the user's global default. Locked decision: per-device, not
// cross-device — view mode is fundamentally shaped by viewport.
// View mode (list / grid / gallery) is a single PER-USER account preference,
// persisted server-side via `users.update` and mirrored locally through
// `authStore.updateUser`. It deliberately retains across folders — the chosen
// layout is the user's, not the folder's, so navigating around never resets it.
// (This replaces the old per-folder localStorage override.)
const persistViewMode = async (mode: ViewModeType) => {
  if (!authStore.user) return;
  const data = { id: authStore.user.id, viewMode: mode };
  try {
    await users.update(data, ["viewMode"]);
  } catch {
    // Failing to persist shouldn't block the visible switch — the optimistic
    // ref update below already applied it for this session.
  }
  authStore.updateUser(data);
};

const setView = async (mode: string) => {
  if (!authStore.user) return;
  if (viewMode.value === mode) return;
  layoutStore.closeHovers();
  viewMode.value = mode as ViewModeType; // optimistic — show immediately
  setItemWeight();
  fillWindow();
  void persistViewMode(mode as ViewModeType);
};

// Source of truth for the active layout. Seeded from the account default and
// kept in sync if that default changes elsewhere (command palette, Profile, or
// another tab via the auth store).
const viewMode = ref<ViewModeType>(authStore.user?.viewMode ?? "list");

watch(
  () => authStore.user?.viewMode,
  (mode) => {
    if (mode && mode !== viewMode.value) viewMode.value = mode;
  }
);

// CH-1: a view-mode switch (list ↔ grid ↔ gallery) changes the tile
// dimensions + column count. Re-measure + re-window AFTER the DOM has
// re-rendered with the new mode's classes (so the grid metrics are read
// from the correct layout, not the outgoing one).
watch(viewMode, () => {
  nextTick(() => {
    listingGrid.measure();
    listingGrid.update();
  });
});

const totalItems = computed(
  () => (fileStore.req?.numDirs ?? 0) + (fileStore.req?.numFiles ?? 0)
);

const allSelected = computed(
  () => totalItems.value > 0 && fileStore.selectedCount === totalItems.value
);

const toggleSelectAll = () => {
  if (allSelected.value) {
    fileStore.selected = [];
  } else {
    const indices = (fileStore.req?.items ?? []).map((it: any) => it.index);
    fileStore.selected = indices;
  }
};

// Bulk-action pill is shown when 2+ items are selected (single selection
// uses InfoPane instead) AND no slide-over panel is open — otherwise the
// fixed bottom pill overlaps the right-hand panel (RC fix).
const showBulkPill = computed(
  () => fileStore.selectedCount >= 2 && !anyPanelOpen.value
);

// Inline new-item row: when the layout store flags "newDir" or "newFile" as
// the current prompt, render the inline row instead of the legacy modal.
// All existing triggers (sidebar New button, palette command, section More
// menu) call `layoutStore.showHover("newDir" | "newFile")`, so this watch
// is the only wiring needed — no caller changes.
const inlineNewKind = computed<"newDir" | "newFile" | null>(() => {
  const name = layoutStore.currentPromptName;
  return name === "newDir" || name === "newFile" ? name : null;
});

// ── Delete → confirm → trash + Undo-restores toast ───────────────────
// (Stage 8 flow, rebuilt for the 2.4.0 Stage 2 recycle bin.) Flow:
//   1. Any "delete" trigger (header button, palette, pill, ctx menu, the
//      Delete key) routes through the confirm dialog.
//   2. Confirm → the delete API runs IMMEDIATELY — the backend MOVES the
//      items into the trash (an instant same-volume rename) and returns a
//      trashId per item.
//   3. An Undo toast shows for 10s; Undo RESTORES the trashed entries (the
//      delete already happened — undo is a real round-trip, so it works
//      even after navigating away). Items also remain recoverable from the
//      Trash view long after the toast is gone.
//   4. Shift+Delete (or the Trash view) deletes permanently: same confirm
//      dialog with "Delete forever" wording, no undo.
// The legacy modal is still kept in Prompts.vue for the file-editor
// delete case (where `isListing === false`).
const $toast = useToast();
const UNDO_WINDOW_MS = 5000;

const confirmOpen = ref(false);
const confirmTitle = ref("");
const confirmMessage = ref("");
const pendingConfirm = ref<{ url: string; name: string }[]>([]);
const pendingPermanent = ref(false);

// After an optimistic delete, move the selection to the nearest remaining
// item so a follow-up Shift+Delete (RC-10) has a target instead of
// no-opping on an empty selection. Mirrors the image-preview delete flow,
// which already advances to a neighbor. Runs BEFORE the deleted items are
// hidden by the pending filter, so the pre-delete visual order is intact.
const selectNeighborAfterDelete = (deletedUrls: Set<string>) => {
  // dirs-then-files matches the listing's render order.
  const visible = [...items.value.dirs, ...items.value.files];
  const firstDeletedPos = visible.findIndex((it) => deletedUrls.has(it.url));
  const remaining = visible.filter((it) => !deletedUrls.has(it.url));
  fileStore.multiple = false;
  if (remaining.length === 0) {
    // Deleted the whole folder's worth of items — nothing left to select.
    fileStore.selected = [];
    return;
  }
  // The item that slides up into the first deleted slot (or the last
  // remaining item if the deletion was at the end of the list).
  const pos = firstDeletedPos === -1 ? 0 : firstDeletedPos;
  const neighbor = remaining[Math.min(pos, remaining.length - 1)];
  // Immediate re-selection — `index` is the same key space the rows bind
  // to, so the selection ring + a follow-up keyboard delete both work now.
  fileStore.selected = [neighbor.index];
  // Survive the eventual reload: performDelete sets reload=true, and
  // Files.vue re-resolves queued preselect paths into the selection.
  if (neighbor.path) fileStore.setPreselect(neighbor.path);
};

// Undo = restore the just-trashed entries by id. A real API round-trip (the
// delete already happened), so it works even after navigating away.
const undoRestore = async (ids: string[]) => {
  try {
    await Promise.all(ids.map((id) => trashApi.restore(id)));
  } catch (e) {
    if (e instanceof Error) $showError(e);
  } finally {
    fileStore.reload = true;
  }
};

const startUndoDelete = async (
  items: { url: string; name: string }[],
  permanent: boolean
) => {
  // Advance the selection to a neighbor before the items vanish (RC-10).
  selectNeighborAfterDelete(new Set(items.map((i) => i.url)));

  if (permanent) {
    try {
      await Promise.all(items.map((i) => api.remove(i.url, true)));
    } catch (e) {
      if (e instanceof Error) $showError(e);
    } finally {
      fileStore.reload = true;
    }
    return;
  }

  // Move to trash NOW (instant same-volume rename server-side), keep the ids
  // for the undo toast. Partial failures: whatever made it into the trash is
  // undoable; the error for the rest surfaces via the toast.
  let trashIds: string[] = [];
  try {
    const results = await Promise.all(items.map((i) => api.remove(i.url)));
    trashIds = results
      .filter((r): r is { trashId: string } => !!r?.trashId)
      .map((r) => r.trashId);
  } catch (e) {
    if (e instanceof Error) $showError(e);
  } finally {
    fileStore.reload = true;
  }
  if (trashIds.length === 0) return;

  const message =
    items.length === 1
      ? `Moved “${items[0].name}” to Trash`
      : `Moved ${items.length} items to Trash`;

  const toastId = $toast(
    {
      component: UndoToast,
      props: {
        message,
        onClick: () => {
          $toast.dismiss(toastId);
          void undoRestore(trashIds);
        },
      },
    },
    {
      timeout: UNDO_WINDOW_MS,
      closeOnClick: false,
      icon: false,
      // Dedicated class so the delete toast gets its own dark-orange skin +
      // width clamp (see .Vue-Toastification__toast.toast--undo in styles.css),
      // distinct from the neutral grey of generic toasts.
      toastClassName: "toast--undo",
    }
  );
};

// Open the trash/permanent confirm for `items`. Shared by the prompt
// intercept watcher (header button, pill, ctx menu, palette) and the
// Delete-key shortcut, so every entry point gets identical wording.
const openDeleteConfirm = (
  items: { url: string; name: string }[],
  permanent: boolean
) => {
  if (items.length === 0) return;
  pendingConfirm.value = items;
  pendingPermanent.value = permanent;
  if (permanent) {
    confirmTitle.value =
      items.length === 1
        ? `Permanently delete “${items[0].name}”?`
        : `Permanently delete ${items.length} items?`;
    confirmMessage.value = "This skips the Trash and cannot be undone.";
  } else {
    const it = items[0];
    const labelHint = it.url.endsWith("/") ? "folder" : "file";
    confirmTitle.value =
      items.length === 1
        ? `Move this ${labelHint} to the Trash?`
        : `Move ${items.length} items to the Trash?`;
    confirmMessage.value =
      items.length === 1
        ? `“${it.name}” can be restored from the Trash later.`
        : "They can be restored from the Trash later.";
  }
  confirmOpen.value = true;
};

const onDeleteConfirm = () => {
  const items = pendingConfirm.value;
  const permanent = pendingPermanent.value;
  confirmOpen.value = false;
  pendingConfirm.value = [];
  pendingPermanent.value = false;
  if (items.length === 0) return;
  void startUndoDelete(items, permanent);
};

const onDeleteCancel = () => {
  confirmOpen.value = false;
  pendingConfirm.value = [];
  pendingPermanent.value = false;
};

// Build the {url, name}[] list for the current listing selection.
// Shared by the confirm-dialog intercept watcher and the S4-5
// skip-confirm shortcut path so both agree on what "the selection" is.
const collectSelectedDeleteItems = (): { url: string; name: string }[] => {
  const req = fileStore.req;
  if (!req) return [];
  return fileStore.selected
    .map((idx) => req.items[idx])
    .filter(Boolean)
    .map((i) => ({ url: i.url, name: i.name }));
};

// ── Stage 8: Move / Copy / Share slide-overs ──────────────────────────
// Same intercept pattern as delete: when one of these prompts fires from
// the listing context, snapshot what we need, dismiss the prompt, and
// open the slide-over panel. The legacy modals remain registered in
// Prompts.vue only as a safety net for non-listing callers.
const moveCopyOpen = ref(false);
const moveCopyMode = ref<"move" | "copy">("move");
// Dual-pane: pane B opens the move/copy picker for ITS selection by passing an
// `override` on the prompt. Captured here (before closeHovers clears the prompt)
// and handed to MoveCopyPanel; null for pane A, which reads fileStore as before.
const moveCopyOverride = ref<{
  items: {
    url: string;
    name: string;
    isDir: boolean;
    size: number;
    modified: string;
  }[];
  sourceUrl: string;
} | null>(null);
const shareOpen = ref(false);
const extractOpen = ref(false);
// v1.3 S4-2: BulkRenamePanel state lives on `bulkRename` (composable
// declared up top with the other singleton composables) so the
// command palette can flip it without prop drilling. Local helpers
// are just thin wrappers that match the SlideOver's cancel/done
// emit shape.

const closeMoveCopy = () => {
  moveCopyOpen.value = false;
  moveCopyOverride.value = null;
};
const closeShare = () => {
  shareOpen.value = false;
};
const closeExtract = () => {
  extractOpen.value = false;
};

// True when ANY of the slide-over panels is open. Drives the pill-hide
// (so the bottom pill doesn't overlap a panel) and the one-at-a-time
// enforcement below.
const anyPanelOpen = computed(
  () =>
    moveCopyOpen.value ||
    shareOpen.value ||
    extractOpen.value ||
    bulkRename.isOpen.value
);

// Only one slide-over may be open at a time. Each open-path calls this
// first so opening (say) Rename dismisses an already-open Copy panel
// instead of stacking them.
const closeAllPanels = () => {
  moveCopyOpen.value = false;
  shareOpen.value = false;
  extractOpen.value = false;
  bulkRename.close();
};

// BulkRename is opened from outside FileListing (pill button, command
// palette, row context menu) via the composable singleton — so close the
// sibling panels reactively whenever it opens.
watch(
  () => bulkRename.isOpen.value,
  (open) => {
    if (open) {
      moveCopyOpen.value = false;
      shareOpen.value = false;
      extractOpen.value = false;
    }
  }
);

watch(
  () => layoutStore.currentPromptName,
  (name) => {
    if (name === "move" || name === "copy") {
      // Pane B passes its own items via `props.override`; capture it BEFORE
      // closeHovers clears the prompt. With an override we skip the pane-A
      // fileStore gate (pane B's selection lives in its own store).
      const override = layoutStore.currentPrompt?.props?.override ?? null;
      if (!override && (!fileStore.isListing || fileStore.selectedCount === 0))
        return;
      moveCopyMode.value = name;
      moveCopyOverride.value = override;
      layoutStore.closeHovers();
      closeAllPanels();
      moveCopyOpen.value = true;
      return;
    }
    if (name === "share") {
      // Share targets a single item — either the file being viewed, or the
      // sole selected item in the listing. If the listing has 0 or 2+
      // selected, do nothing (legacy modal would have noop'd too).
      const singleListingSelection =
        fileStore.isListing && fileStore.selectedCount === 1;
      const fileView = !fileStore.isListing;
      if (!singleListingSelection && !fileView) return;
      layoutStore.closeHovers();
      closeAllPanels();
      shareOpen.value = true;
      return;
    }
    if (name === "extract") {
      // Extract targets one OR MORE selected archives in the current listing
      // (each extracts into its own subfolder). Mirror the gate from
      // `headerButtons.extract` so a stray `layoutStore.showHover("extract")`
      // from a stale code path can't open the panel with no valid source:
      // every selected item must be extractable.
      if (!unzipEnabled) return;
      const req = fileStore.req;
      if (!fileStore.isListing || fileStore.selectedCount < 1 || !req) return;
      const allArchives = fileStore.selected.every((i) =>
        isExtractable(req.items[i]?.name ?? "")
      );
      if (!allArchives) return;
      layoutStore.closeHovers();
      closeAllPanels();
      extractOpen.value = true;
      return;
    }
  }
);

watch(
  () => layoutStore.currentPromptName,
  (name) => {
    if (name !== "delete") return;
    // Only intercept when we have a listing-level selection; the legacy
    // modal still handles file-editor deletes (a different code path).
    if (!fileStore.isListing || fileStore.selectedCount === 0) return;

    const req = fileStore.req;
    if (!req) return;
    const items = collectSelectedDeleteItems();
    if (items.length === 0) return;

    // Dismiss the prompt so the legacy modal doesn't render alongside ours
    layoutStore.closeHovers();

    openDeleteConfirm(items, false);
  }
);

const clearSelection = () => {
  fileStore.selected = [];
  fileStore.multiple = false;
};

// Section title "More" dropdown
const sectionMoreShow = ref<boolean>(false);
const sectionMorePos = ref<{ x: number; y: number }>({ x: 0, y: 0 });

const showSectionMore = (event: MouseEvent) => {
  const btn = event.currentTarget as HTMLElement;
  const rect = btn.getBoundingClientRect();
  sectionMoreShow.value = true;
  // Right-align menu's right edge to the button's right edge (menu min-width is 200px)
  sectionMorePos.value = {
    x: rect.right - 200,
    y: rect.bottom + 6 + window.scrollY,
  };
};

const hideSectionMore = () => {
  sectionMoreShow.value = false;
};

// Items array for the items-based ContextMenu (v1.3.0 S1-3). Built as
// a computed so per-render permission checks (e.g., canRenameCurrentFolder
// reactive on route) flow through naturally. `.filter(Boolean)` drops
// items gated off by permissions — TS narrows the result back to the
// non-null MenuItem[] via the explicit return type.
const sectionMoreItems = computed<MenuItem[]>(() => {
  const canCreate = !!authStore.user?.perm.create;
  const items: (MenuItem | null)[] = [
    canCreate
      ? {
          label: t("sidebar.newFolder"),
          icon: "folder-plus",
          action: () => layoutStore.showHover("newDir"),
        }
      : null,
    canCreate
      ? {
          label: t("sidebar.newFile"),
          icon: "file-plus",
          action: () => layoutStore.showHover("newFile"),
        }
      : null,
    // Rename the CURRENTLY VIEWED folder (not selected items).
    // Hidden at the storage root since you can't rename "/". Also hidden while
    // split: the inline rename input lives in the hero, which is swapped for the
    // compact header in split view, so there'd be nowhere to type (silent no-op).
    canRenameCurrentFolder.value && !splitActive.value
      ? {
          label: "Rename folder",
          icon: "pencil",
          action: startFolderRename,
        }
      : null,
    // Edit the sidebar Favorites alias for the CURRENT folder — only when it's
    // pinned (the title is meaningless otherwise). Sidebar-only; the real
    // folder is untouched.
    currentFolderFavorited.value
      ? {
          label: "Favorites display title…",
          icon: "star",
          action: () => favTitleDialog.open(currentFolderPath.value),
        }
      : null,
    // Selection actions — each appears ONLY when it's valid for the current
    // selection (same gates the InfoPane rail uses), so the ⋯ menu mirrors
    // exactly what you can do with what's selected: Extract shows for a single
    // archive but not an image, Edit tags shows only when audio is selected,
    // etc. (V2-J: Share + Download were moved off the hero cluster into ⋯.)
    headerButtons.value.share
      ? {
          label: t("buttons.share"),
          icon: "share",
          action: () => layoutStore.showHover("share"),
        }
      : null,
    headerButtons.value.download
      ? {
          label: t("buttons.download"),
          icon: "download",
          action: download,
        }
      : null,
    headerButtons.value.extract
      ? {
          label: "Extract",
          icon: "package-open",
          action: () => layoutStore.showHover("extract"),
        }
      : null,
    bulkAudioCount.value >= 1 && !!authStore.user?.perm.modify
      ? {
          label: "Edit tags",
          icon: "music",
          action: () => layoutStore.showHover("audio-tags"),
        }
      : null,
    headerButtons.value.rename
      ? {
          label: t("buttons.rename"),
          icon: "pencil",
          action: () => layoutStore.showHover("rename"),
        }
      : null,
    headerButtons.value.copy
      ? {
          label: t("buttons.copyFile"),
          icon: "copy",
          action: () => layoutStore.showHover("copy"),
        }
      : null,
    headerButtons.value.move
      ? {
          label: t("buttons.moveFile"),
          icon: "forward",
          action: () => layoutStore.showHover("move"),
        }
      : null,
    headerButtons.value.delete
      ? {
          label: t("buttons.delete"),
          icon: "trash-2",
          action: () => layoutStore.showHover("delete"),
        }
      : null,
    { label: "Refresh", icon: "rotate-ccw", action: refresh, kbd: "/" },
    {
      label: t("buttons.info"),
      icon: "info",
      action: () => layoutStore.showHover("info"),
    },
  ];
  return items.filter((i): i is MenuItem => i !== null);
});

const refresh = () => {
  fileStore.reload = true;
  // Dual-pane: also refresh pane B so a refresh updates BOTH panes (especially
  // when they're showing the same folder). Cheap re-fetch; no-op visually if its
  // folder didn't change.
  if (splitActive.value) panes.refreshB();
};

// Pane A's split close button (X): close the split and make pane B's folder the
// single open folder — i.e. close the pane the user clicked, keep the other. (The
// generic toggle just drops the split keeping pane A; that's the wrong pane when
// closing FROM pane A's header.) Pane B's own X already keeps pane A via closeSplit().
const closeSplitToB = () => {
  const target = panes.secondaryPath;
  panes.closeSplit();
  if (target && target !== route.path) void router.push({ path: target });
};

// ── S7-1: pull-to-refresh (touch only) ──────────────────────────────
// The scroll container differs by view (S6-1): list view scrolls inside
// the RecycleScroller, grid/gallery scroll the <section>. Resolve the
// active one reactively; the composable re-binds its listeners when this
// swaps on a view-mode change.
const ptrScrollEl = computed<HTMLElement | null>(() => {
  if (isListVirtual.value) {
    const inst = listScroller.value as unknown as { $el?: HTMLElement } | null;
    return inst?.$el ?? null;
  }
  return scrollSection.value;
});

// Reload the folder and hold the spinner until the fetch settles. A
// reload flips layoutStore.loading true→false (Files.vue fetchData), so
// we resolve on that transition — with a small floor so the spinner
// doesn't flash, and a safety timeout in case loading never toggles.
const onPullRefresh = () =>
  new Promise<void>((resolve) => {
    const startedAt = Date.now();
    let sawLoading = false;
    const settle = () => {
      stop();
      clearTimeout(safety);
      setTimeout(resolve, Math.max(0, 450 - (Date.now() - startedAt)));
    };
    const stop = watch(
      () => layoutStore.loading,
      (loading) => {
        if (loading) sawLoading = true;
        else if (sawLoading) settle();
      }
    );
    // Safety net: resolve anyway if a reload never toggles `loading`.
    const safety = setTimeout(() => {
      stop();
      resolve();
    }, 5000);
    refresh();
  });

const {
  active: ptrActive,
  offset: ptrOffset,
  opacity: ptrOpacity,
  rotation: ptrRotation,
  refreshing: ptrRefreshing,
  pulling: ptrPulling,
} = usePullToRefresh({
  el: ptrScrollEl,
  threshold: 64,
  onRefresh: onPullRefresh,
  // Suppressed unless we're on a touch device with nothing else going on:
  // not already loading, no open prompt, no active selection / drag /
  // lasso (those own the gesture surface).
  enabled: () =>
    isTouchDevice.value &&
    !layoutStore.loading &&
    layoutStore.currentPrompt === null &&
    fileStore.selectedCount === 0 &&
    fileStore.draggedItems.length === 0 &&
    dragSelect.lasso.value === null,
});

const sortLabel = computed(() => {
  const by = fileStore.req?.sorting.by ?? "name";
  if (by === "name") return t("files.name");
  if (by === "size") return t("files.size");
  // "Modified" instead of "Last modified" — the longer string overflowed
  // the header sort button at near-md widths once the chevron was removed.
  if (by === "modified") return "Modified";
  // v1.3 S3-5: extension sort label. "Type" is shorter than
  // "Extension" and reads naturally in the cycle button at narrow
  // widths without needing min-width adjustment.
  if (by === "extension") return "Type";
  return by;
});

// Human-readable sort direction, shown in the Sort button's tooltip and the
// popover's Direction rows. Direction persists per-user via user.sorting.asc
// (server-side), so it sticks across folders until changed. Changing it now
// lives inside the consolidated Sort popover (see sortMenuItems) rather than a
// separate toolbar toggle button.
const sortDirLabel = computed(() =>
  ascOrdered.value ? "Ascending" : "Descending"
);

// ── Multi-column sort popover (v1.3 S3-4) ───────────────────────────
// Replaces the legacy cycle button (and the brief S3-5-extension
// cycle extension). Click → ContextMenu with primary + secondary
// criterion selection. Primary persists server-side via the existing
// users.update flow; secondary persists client-side via
// usePreferences and applies as an in-memory tiebreaker after fetch.

const SORT_OPTIONS: Array<{ key: SortKey; label: string }> = [
  { key: "name", label: t("files.name") },
  { key: "size", label: t("files.size") },
  { key: "modified", label: "Modified" },
  { key: "extension", label: "Type" },
];

// ── View menu (v2.7.2 header declutter) ────────────────────────────
// One popover replaces the 3-button layout switcher + the split toggle —
// the cluster read as a crowd. Same anchoring + toggle-on-reclick
// conventions as the Sort popover below; a check marks the active choice.
const VIEW_OPTIONS = [
  { key: "list", label: "List", icon: "list" },
  { key: "mosaic", label: "Grid", icon: "layout-grid" },
  { key: "mosaic gallery", label: "Gallery", icon: "image" },
] as const;

const viewMenuShow = ref(false);
const viewMenuPos = ref({ x: 0, y: 0 });

const viewModeIcon = computed(
  () => VIEW_OPTIONS.find((o) => o.key === viewMode.value)?.icon ?? "list"
);
const viewModeLabel = computed(
  () => VIEW_OPTIONS.find((o) => o.key === viewMode.value)?.label ?? "List"
);

const openViewMenu = (event: MouseEvent) => {
  if (viewMenuShow.value) {
    viewMenuShow.value = false;
    return;
  }
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  viewMenuPos.value = { x: rect.left, y: rect.bottom + 4 };
  viewMenuShow.value = true;
};

const viewMenuItems = computed<MenuItem[]>(() => {
  const items: MenuItem[] = [
    { type: "header", label: "Layout" },
    ...VIEW_OPTIONS.map((o) => ({
      label: o.label,
      icon: viewMode.value === o.key ? "check" : o.icon,
      action: () => void setView(o.key),
    })),
  ];
  if (splitAvailable.value) {
    items.push(
      { type: "separator" },
      {
        label: splitActive.value ? "Close split view" : "Split view",
        icon: splitActive.value ? "check" : "columns-2",
        action: () => toggleSplit(),
      }
    );
  }
  return items;
});

const sortMenuShow = ref(false);
const sortMenuPos = ref<{ x: number; y: number }>({ x: 0, y: 0 });

const openSortMenu = (event: MouseEvent) => {
  // V3-C #7: clicking the trigger again closes an open menu (toggle) instead of
  // forcing a click outside. `.stop` on the trigger hides this click from
  // ContextMenu's outside-click listener, so we have to toggle explicitly.
  if (sortMenuShow.value) {
    sortMenuShow.value = false;
    return;
  }
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  // Anchor the popover to the bottom-left corner of the button so it
  // reads as "belongs to the button." ContextMenu's positioner clamps
  // to viewport, so right-edge overflow is auto-handled.
  sortMenuPos.value = { x: rect.left, y: rect.bottom + 4 };
  sortMenuShow.value = true;
};

/** Read the saved secondary criterion (if any). Stored in prefs as a
 *  SortCriterion or null. */
const secondarySort = computed<SortCriterion | null>(() =>
  prefs.get<SortCriterion | null>("sort.secondary", null)
);

/** Apply the user's choice to the PRIMARY axis. Same flow as before:
 *  PUT to user.sorting + trigger a reload. */
const setPrimarySort = (by: SortKey, asc: boolean) => {
  sortMenuShow.value = false;
  void sortRaw(by, asc);
};

/** Update the SECONDARY axis. Stored in prefs; no server round-trip;
 *  no reload — the items computed re-applies the tiebreaker reactively. */
const setSecondarySort = (criterion: SortCriterion | null) => {
  sortMenuShow.value = false;
  void prefs.set("sort.secondary", criterion);
};

/** Build the ContextMenu items array — the single consolidated Sort popover
 *  (field + direction + secondary, replacing the old two-button pair). Layout:
 *
 *    PRIMARY                ← header
 *    Name           [check] ← active field is checked
 *    Size
 *    Modified
 *    Type
 *    ────────────           ← separator
 *    DIRECTION              ← header
 *    Ascending      [check] ← active direction is checked
 *    Descending
 *    ────────────           ← separator
 *    THEN BY                ← header
 *    None           [check]
 *    Name           [arrow] ← secondary shows its own direction arrow
 *    Size
 *    Modified
 *    Type
 *
 * Picking a primary field keeps the current direction; the Direction rows
 * set it explicitly. Picking an inactive secondary sets it to ascending;
 * re-picking it flips direction. "None" clears the secondary. */
const sortMenuItems = computed<MenuItem[]>(() => {
  const primaryBy = (fileStore.req?.sorting.by ?? "name") as SortKey;
  const primaryAsc = fileStore.req?.sorting.asc ?? false;
  const sec = secondarySort.value;

  const items: MenuItem[] = [
    { type: "header", label: "Primary" },
    ...SORT_OPTIONS.map((opt) => ({
      label: opt.label,
      // A check marks the active field. Picking a field keeps the current
      // direction; direction is set in the Direction section below.
      icon: primaryBy === opt.key ? "check" : undefined,
      action: () => {
        setPrimarySort(opt.key, primaryAsc);
      },
    })),
    { type: "separator" },
    { type: "header", label: "Direction" },
    {
      label: "Ascending",
      icon: primaryAsc ? "check" : undefined,
      action: () => setPrimarySort(primaryBy, true),
    },
    {
      label: "Descending",
      icon: !primaryAsc ? "check" : undefined,
      action: () => setPrimarySort(primaryBy, false),
    },
    { type: "separator" },
    { type: "header", label: "Then by" },
    {
      label: "None",
      icon: sec === null ? "check" : undefined,
      action: () => setSecondarySort(null),
    },
    ...SORT_OPTIONS.map((opt) => ({
      // Disable picking the same key for primary + secondary — that
      // would be a no-op tiebreaker (every primary tie also ties on
      // the same key). Greyed out so the rule is visible.
      label: opt.label,
      disabled: opt.key === primaryBy,
      icon:
        sec?.by === opt.key ? (sec.asc ? "arrow-up" : "arrow-down") : undefined,
      action: () => {
        const nextAsc = sec?.by === opt.key ? !sec.asc : true;
        setSecondarySort({ by: opt.key, asc: nextAsc });
      },
    })),
  ];
  return items;
});

/** Pure sort dispatcher — used by the popover. The legacy `sort()`
 *  function still exists below (called by column-header clicks) and
 *  reuses this codepath internally so the persistence story is
 *  centralized. */
const sortRaw = async (by: SortKey, asc: boolean) => {
  // Optimistic + client-authoritative: update the in-memory sorting NOW so the
  // listing and the sort icons re-order this frame (the `items` computed sorts
  // client-side). Then persist to the server so the choice sticks on the next
  // fresh load. No forced reload — the client already shows the new order, and
  // relying on a silent background reload to re-sort was the source of the
  // "sort button does nothing" bug.
  const prev = fileStore.req?.sorting;
  if (fileStore.req) fileStore.req.sorting = { by, asc };
  try {
    if (authStore.user?.id) {
      await users.update({ id: authStore.user?.id, sorting: { by, asc } }, [
        "sorting",
      ]);
    }
    // Race guard: a silent background refresh (transfer/upload/tag tick) that
    // landed WHILE the PUT was in flight calls updateRequest(), which swaps in
    // the server's PRE-update sorting and snaps the list back. The PUT has now
    // committed our value server-side, so if the current sorting is still that
    // stale pre-change value, re-assert ours locally (no reload) to win the
    // race. If it's a *different* value, the user picked another sort in the
    // meantime — leave their newer choice alone.
    const cur = fileStore.req?.sorting;
    if (
      fileStore.req &&
      cur &&
      prev &&
      cur.by === prev.by &&
      cur.asc === prev.asc &&
      (cur.by !== by || cur.asc !== asc)
    ) {
      fileStore.req.sorting = { by, asc };
    }
  } catch (e: any) {
    // Roll the optimistic order back so the view doesn't show a sort the
    // server never accepted.
    if (fileStore.req && prev) fileStore.req.sorting = prev;
    $showError(e);
  }
};

const uploadFunc = () => {
  if (
    typeof window.DataTransferItem !== "undefined" &&
    typeof DataTransferItem.prototype.webkitGetAsEntry !== "undefined"
  ) {
    layoutStore.showHover("upload");
  } else {
    document.getElementById("upload-input")?.click();
  }
};

const setItemWeight = () => {
  // CH-1: legacy hook from the old showLimit windowing. Now it just
  // re-measures the grid (columns + tile height). Kept (rather than ripped
  // out of every caller) so the call sites stay readable.
  if (isListVirtual.value) return;
  listingGrid.measure();
};

const fillWindow = () => {
  // CH-1: legacy hook — recompute the grid/gallery visible window. The
  // virtualized list view manages its own window inside the recycler.
  if (isListVirtual.value) return;
  listingGrid.update();
};

const revealPreviousItem = () => {
  if (!fileStore.req || !fileStore.oldReq) return;

  const index = fileStore.selected[0];
  if (index === undefined) return;

  // CH-1: the target tile may be windowed out of the DOM, so scroll the
  // section to its computed row first, then let the scroll handler mount
  // it. Locate the item's position within its section (folders or files),
  // compute the content-space Y, and center it in the viewport.
  const sc = scrollSection.value;
  if (!sc) return;
  const { dirs: dItems, files: fItems } = items.value;
  let ordinal = dItems.findIndex((it) => it.index === index);
  let sectionEl = dirsSectionRef.value;
  if (ordinal < 0) {
    ordinal = fItems.findIndex((it) => it.index === index);
    sectionEl = filesSectionRef.value;
  }
  if (ordinal < 0) return;

  nextTick(() => {
    // Gallery folders use a narrower (smaller-tile) grid than files, so pull
    // the metrics for whichever section this item lives in.
    const isDirSec = sectionEl === dirsSectionRef.value;
    const cols =
      (isDirSec ? listingGrid.dirsCols.value : listingGrid.filesCols.value) ||
      1;
    const stride =
      (isDirSec ? listingGrid.dirsTileH.value : listingGrid.filesTileH.value) +
      12;
    const row = Math.floor(ordinal / cols);
    const secTop = sectionEl
      ? sectionEl.getBoundingClientRect().top -
        sc.getBoundingClientRect().top +
        sc.scrollTop
      : 0;
    const target = secTop + row * stride - sc.clientHeight / 2;
    sc.scrollTop = Math.max(0, target);
    listingGrid.update();
  });

  return true;
};

// ── S4-1 context menu dispatch ──────────────────────────────────────
// Single entry point invoked by `#listing`'s @contextmenu. We decide
// which menu (row vs background) based on whether the event target sits
// inside an `.item` element. Both menus reuse the same `<context-menu>`
// instance — we just hand it a different `:items` array.
//
// Row context: each ListingItem's own @contextmenu has already fired
// FIRST (events bubble inside-out) and adopted/replaced the selection
// per the locked spec (right-click on unselected = select that row;
// right-click on already-selected = preserve multi). All we do here
// is open the menu with the current selection-derived items.
//
// Background context: emit a different items array (New folder, etc.).
const contextMenuMode = ref<"row" | "background">("row");

const onListingContextMenu = (event: MouseEvent) => {
  // macOS synthesizes a `contextmenu` event from ctrl+left-click. Treat that
  // as a (multi-)select modifier, NOT a right-click: suppress both the custom
  // and the native menu so ctrl+drag can lasso and ctrl+click just toggles
  // selection (matching cmd+click) instead of popping a menu on first click.
  if (event.ctrlKey) {
    event.preventDefault();
    return;
  }
  event.preventDefault();
  const target = event.target as HTMLElement | null;
  // Treat `.item.header` (the column-header row) as background — it
  // shares the `.item` class for layout grid reasons but isn't a
  // selectable file row. Without this, right-clicking the header
  // would put us in "row" mode with an empty selection and the menu
  // would open with no items.
  const itemEl = target?.closest?.(".item") as HTMLElement | null;
  const insideRealItem = itemEl != null && !itemEl.classList.contains("header");
  contextMenuMode.value = insideRealItem ? "row" : "background";
  isContextMenuVisible.value = true;
  contextMenuPos.value = {
    x: event.clientX + 8,
    y: event.clientY + Math.floor(window.scrollY),
  };
};

const hideContextMenu = () => {
  isContextMenuVisible.value = false;
};

// ── S4-1 row context menu items ─────────────────────────────────────
// Computed `MenuItem[]` for right-clicks landing on a row. Reuses
// `headerButtons` for the permission/selection-size/file-type gating
// the bulk pill already does, then adds the S4-1 additions (Open,
// Tag, Copy path). Items that don't pass their gate aren't emitted at
// all — the ContextMenu primitive handles separators between groups
// even when leading items disappear.
//
// `Open` is single-selection only (multi-open opens N tabs which we
// don't want — would also race the route guard). `Rename` is single
// too (bulk rename is S4-2). `Move` / `Copy` / `Delete` / `Download`
// work on the full selection. `Extract` is single .zip only (gated
// inside headerButtons.extract). `Share` is single-only too.
const rowMenuItems = computed<MenuItem[]>(() => {
  const sel = fileStore.selectedCount;
  if (sel === 0) return []; // shouldn't open the menu with no selection
  const items: MenuItem[] = [];
  const hb = headerButtons.value;
  const singleItem =
    sel === 1 ? (fileStore.req?.items[fileStore.selected[0]] ?? null) : null;

  // ── Open / Tag / Copy path (single-selection actions) ────────────
  if (singleItem) {
    items.push({
      label: singleItem.isDir ? "Open folder" : "Open",
      icon: "external-link",
      action: () => {
        hideContextMenu();
        if (singleItem) void router.push({ path: singleItem.url });
      },
    });
  }
  // Open a folder in a SECOND pane (dual-pane). Single-pane only — once split is
  // already on, the cross-pane "Move/Copy to other pane" actions cover the rest.
  if (singleItem?.isDir && !splitActive.value && splitAvailable.value) {
    const dirUrl = singleItem.url;
    items.push({
      label: "Open in new pane",
      icon: "columns-2",
      action: () => {
        hideContextMenu();
        panes.openSplit(route.path.replace(/\/?$/, "/"));
        void panes.navigateB(dirUrl);
      },
    });
  }
  if (singleItem) {
    items.push({
      label: "Tag…",
      icon: "tag",
      action: () => {
        hideContextMenu();
        // Route through the bulk tag sheet (mounted at this component's top
        // level, always available) instead of `tagPicker` — that one lives
        // INSIDE InfoPane behind `v-if="item"`, so it silently no-ops when the
        // details rail is collapsed or has no item. A single-path array gives
        // the same single-file tag UX.
        bulkTagPicker.open([singleItem.url]);
      },
    });
  }
  // Favorites display title — only for a folder that's currently pinned, since
  // the title only shows in the sidebar Favorites section. Sets a sidebar-only
  // alias; never renames the real folder.
  if (
    singleItem &&
    singleItem.isDir &&
    favoritesComposable.isFavorited(singleItem.url)
  ) {
    const favPath = singleItem.url;
    items.push({
      label: "Favorites display title…",
      icon: "star",
      action: () => {
        hideContextMenu();
        favTitleDialog.open(favPath);
      },
    });
  }

  // ── Share / Extract (single-only, gated by perms + file type) ────
  if (hb.share) {
    items.push({
      label: t("buttons.share"),
      icon: "share",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("share");
      },
    });
  }
  if (hb.extract) {
    items.push({
      label: t("buttons.unzip"),
      icon: "package-open",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("extract");
      },
    });
  }
  // Edit audio tags (1.6.0) — single MP3 / FLAC file, requires modify perm.
  if (
    singleItem &&
    isAudioTaggable(singleItem.name) &&
    authStore.user?.perm.modify
  ) {
    items.push({
      label: "Edit tags…",
      icon: "music",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("audio-tags");
      },
    });
  } else if (canBulkEditTags.value) {
    // Batch variant — multiple audio files in the selection. The editor opens
    // blank and applies only the fields the user changes to all of them.
    items.push({
      label: `Edit tags on ${bulkAudioCount.value} files…`,
      icon: "music",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("audio-tags");
      },
    });
  }

  // ── Clipboard: Cut / Copy / Paste into folder (2.4.0 Stage 1) ────
  // Same gates as the Move/Copy pickers (cut pastes as a MOVE → perm.rename;
  // copy pastes as a COPY → perm.create). "Paste into folder" appears on a
  // single selected folder when the clipboard is armed — pasting INTO it
  // without navigating first.
  if (items.length > 0) items.push({ type: "separator" });
  if (hb.move) {
    items.push({
      label: sel === 1 ? "Cut" : `Cut ${sel} items`,
      icon: "scissors",
      kbd: "⌘X",
      action: () => {
        hideContextMenu();
        clipboardCapture("cut");
      },
    });
  }
  if (hb.copy) {
    items.push({
      label: sel === 1 ? "Copy" : `Copy ${sel} items`,
      icon: "copy",
      kbd: "⌘C",
      action: () => {
        hideContextMenu();
        clipboardCapture("copy");
      },
    });
  }
  if (singleItem?.isDir && clipboardStore.items.length > 0) {
    const folderUrl = singleItem.url;
    items.push({
      label: "Paste into folder",
      icon: "clipboard",
      action: () => {
        hideContextMenu();
        void paste(folderUrl);
      },
    });
  }

  // ── Rename / Move to / Copy to / Copy path / Download (varies) ───
  // (ContextMenu renders every separator literally, so don't stack one on an
  // empty or already-separated tail — e.g. a read-only user's menu.)
  if (items.length > 0 && items[items.length - 1].type !== "separator") {
    items.push({ type: "separator" });
  }
  // Single-selection rename → existing inline rename flow.
  // Multi-selection rename → S4-2 BulkRenamePanel (different code path,
  // entirely different UX). Permission gate is the same (perm.rename).
  if (hb.rename) {
    items.push({
      label: t("buttons.rename"),
      icon: "pencil",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("rename");
      },
    });
  } else if (authStore.user?.perm.rename && sel > 1) {
    items.push({
      label: `Bulk rename ${sel} items…`,
      icon: "pencil",
      action: () => {
        hideContextMenu();
        bulkRename.open();
      },
    });
  }
  // "… to…" labels = the destination-picker slide-overs, distinct from the
  // clipboard Cut/Copy entries above (Stage 1) which act via paste.
  if (hb.move) {
    items.push({
      // Folder-aware single-item label so a folder reads "Move folder to…",
      // not "Move file to…". Multi-select uses the neutral "N items" wording.
      label:
        sel === 1
          ? singleItem?.isDir
            ? "Move folder to…"
            : "Move file to…"
          : `Move ${sel} items to…`,
      icon: "forward",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("move");
      },
    });
  }
  if (hb.copy) {
    items.push({
      label:
        sel === 1
          ? singleItem?.isDir
            ? "Copy folder to…"
            : "Copy file to…"
          : `Copy ${sel} items to…`,
      icon: "copy-plus",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("copy");
      },
    });
  }
  if (singleItem) {
    items.push({
      label: "Copy path",
      icon: "link",
      kbd: "⌘⇧C",
      action: () => {
        hideContextMenu();
        void copyItemPath(singleItem);
      },
    });
  }
  if (hb.download) {
    items.push({
      label:
        sel === 1
          ? t("buttons.download")
          : `${t("buttons.download")} ${sel} items`,
      icon: "download",
      action: () => {
        hideContextMenu();
        download();
      },
    });
  }

  // ── Delete (destructive — visual separation + red tint) ──────────
  if (hb.delete) {
    items.push({ type: "separator" });
    items.push({
      label: sel === 1 ? t("buttons.delete") : `Delete ${sel} items`,
      icon: "trash-2",
      destructive: true,
      kbd: "Del",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("delete");
      },
    });
  }

  return items;
});

// ── S4-1 background context menu items ──────────────────────────────
// Shown when right-click lands on empty listing space (gaps between
// rows, the area below the last row, the area beneath the
// folder-file divider). Distinct intent from the row menu — these are
// "act on the current folder" not "act on a selection."
const backgroundMenuItems = computed<MenuItem[]>(() => {
  const items: MenuItem[] = [];
  const canCreate = !!authStore.user?.perm.create;

  if (canCreate) {
    items.push({
      label: "New folder",
      icon: "folder-plus",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("newDir");
      },
    });
    items.push({
      label: "New file",
      icon: "file-plus",
      action: () => {
        hideContextMenu();
        layoutStore.showHover("newFile");
      },
    });
    items.push({
      label: t("buttons.upload"),
      icon: "upload",
      action: () => {
        hideContextMenu();
        uploadFunc();
      },
    });
  }

  // Paste only when the clipboard store has cut/copy contents (placed by
  // ⌘X / ⌘C or the row menu's Cut / Copy). The `paste()` helper handles
  // conflict resolution + clipboard reset for cut.
  if (clipboardStore.items.length > 0) {
    items.push({
      label: "Paste",
      icon: "clipboard",
      kbd: "⌘V",
      action: () => {
        hideContextMenu();
        void paste();
      },
    });
  }

  if (items.length > 0) items.push({ type: "separator" });

  // "Sort by…" hops to the existing sort popover. The submenu support
  // ContextMenu doesn't currently have means we re-anchor the sort
  // menu where the context menu opened (close-enough), so the user
  // sees a continuous popover flow rather than a sudden teleport.
  items.push({
    label: "Sort by…",
    icon: "arrow-down-narrow-wide",
    action: () => {
      const { x, y } = contextMenuPos.value;
      hideContextMenu();
      sortMenuPos.value = { x, y };
      sortMenuShow.value = true;
    },
  });

  items.push({
    label: "Refresh",
    icon: "rotate-ccw",
    kbd: "/",
    action: () => {
      hideContextMenu();
      fileStore.reload = true;
    },
  });

  return items;
});

/** Active items array passed to the shared `<context-menu>` instance.
 *  Switches on `contextMenuMode`, set by `onListingContextMenu`. */
const contextMenuItems = computed<MenuItem[]>(() =>
  contextMenuMode.value === "row"
    ? rowMenuItems.value
    : backgroundMenuItems.value
);

/** Copy the given item's path to the system clipboard. Mirrors the
 *  InfoPane copy-path behavior (toast on success, silent on total
 *  failure) but uses the shared `copy({ text })` util from
 *  `@/utils/clipboard` which already handles secure-context fallback. */
const copyItemPath = async (item: ResourceItem) => {
  const path = item.path || item.url;
  if (!path) return;
  try {
    await copy({ text: path });
    $showSuccess(`Path copied: ${path}`);
  } catch {
    /* swallow — copy() already tried both APIs; nothing more to do */
  }
};

// Whether the press that produced the pending click STARTED on empty listing
// space (vs inside a row or the inline rename input). Recorded in the capture
// phase so it's set before dragSelect's bubble-phase handler runs.
let listingPressStartedOnEmpty = false;
const recordListingPressOrigin = (e: MouseEvent) => {
  listingPressStartedOnEmpty = pressStartedOnEmptyListing(e.target);
};

const handleEmptyAreaClick = (e: MouseEvent) => {
  const target = e.target;
  if (!(target instanceof HTMLElement)) return;

  // Only clear the selection on a GENUINE in-place click on empty space — one
  // whose press also STARTED on empty space. A drag that begins inside a row
  // or the rename input and releases on empty space fires a `click` on the
  // nearest common ancestor (this clear-on-click container); without this
  // guard that wiped the selection, which unmounted an in-progress rename
  // input (v-if="isRenaming") and destroyed the text the user was selecting.
  if (!listingPressStartedOnEmpty) return;

  if (target.dataset.clearOnClick === "true") {
    fileStore.selected = [];
  }
};
</script>
<style scoped>
#listing {
  /* 2.1 #1: fill the scroll section (now min-h-0-bounded) as a flex child,
     instead of a wrong `100vh - 8rem` constant. That constant predated the
     bottom breadcrumb bar and oversized the listing so its last rows ran under
     the bar; flex-fill leaves only the intended ~5px gap above the bar and lets
     the section's overflow scroll engage. Grid/gallery grow with content and
     the section scrolls; list view's RecycleScroller (.listing--virtual) keeps
     its own identical flex sizing below. */
  flex: 1 1 auto;
  min-height: 0;
}

/* Colorful download button in the folder header toolbar — blue-tinted so the
   "save this out" action reads as a real action next to the neutral Share /
   More buttons rather than flat grey chrome. */
.listing-download {
  color: var(--c-blue);
  border-color: color-mix(in srgb, var(--c-blue) 30%, var(--color-line));
  background: color-mix(in srgb, var(--c-blue) 10%, var(--color-surface));
}
.listing-download:hover {
  background: color-mix(in srgb, var(--c-blue) 18%, var(--color-surface));
}

/* ── List virtualization layout (v1.3 S6-1) ──────────────────────────
   In list view the rows live inside a RecycleScroller that owns the
   scroll. The surrounding <section> must NOT also scroll (that would
   double-scroll), and #listing becomes a flex column so the recycler can
   flex to fill the height beneath the now-pinned column header. All of
   this is gated behind the virtual-mode flag classes, so grid + gallery
   keep their native section scroll + showLimit windowing untouched. */
.scroll-section--virtual {
  overflow: hidden;
}

/* Cross-pane drop affordance for pane A (#17) — the same dashed accent frame
   ComparePane draws (`.compare-body--drop`), so dragging a selection over the
   primary pane lights it up identically. Split-only (the flag is gated on
   `splitActive`), so single-pane drag never shows this. */
.scroll-section--drop {
  outline: 2px dashed var(--color-accent, #6e72d9);
  outline-offset: -4px;
  border-radius: 8px;
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.06));
}

#listing.listing--virtual {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  /* Override the tall default min-height (id-specificity) so the column
     can actually shrink to its flex track and hand the recycler a
     bounded height to virtualize against. */
  min-height: 0;
  /* The global `#listing { padding: 0 1rem 1rem }` adds a 1rem bottom pad that
     suits grid/gallery (a margin under the last tile). But here #listing is a
     flex COLUMN whose RecycleScroller child flex-fills — so that bottom pad
     lands BELOW the scroller, shrinking its viewport and lifting the clip point
     ~16px off the breadcrumb (a persistent dead strip above the bar). Zero it
     so the scroller reaches the section's bottom edge; the breadcrumb bar's own
     5px top padding supplies the intended slim gap. Side padding is untouched
     (rows keep their 1rem inset). */
  padding-bottom: 0;
}

.listing-virtual {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  /* Keep row width steady as the scrollbar appears; on macOS overlay
     scrollbars this reserves 0px so the header + rows stay pixel-aligned. */
  scrollbar-gutter: stable;
}

/* The recycler wraps each row in an absolutely-positioned item-view (its
   own BFC). Drop the row's 2px margins there and let the fixed 44px
   item-size supply the rhythm: a 40px row sits atop a 4px gap — visually
   identical to the non-virtual list, but deterministic regardless of
   margin-collapse quirks. */
#listing.list :deep(.vue-recycle-scroller__item-view > .item) {
  margin: 0;
}

/* ── Pull-to-refresh (v1.3 S7-1) ─────────────────────────────────────
   The <section> hosts the indicator; it's only ever shown mid-gesture,
   when scrollTop is pinned at 0, so absolute top:0 sits at the visible
   top. Tracks the finger 1:1 while pulling (no transition); animates the
   snap-back / hold once released. */
.ptr-host {
  position: relative;
}

.ptr-indicator {
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 30;
  pointer-events: none;
  transition:
    transform 0.22s ease,
    opacity 0.22s ease;
}
.ptr-indicator--pulling {
  transition: none;
}

.ptr-indicator__pill {
  margin-top: 8px;
  width: 34px;
  height: 34px;
  border-radius: var(--radius-full, 9999px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-surface, #fff);
  color: var(--color-accent, #6e72d9);
  border: 1px solid var(--color-line, #ececec);
  box-shadow: var(--shadow-md, 0 4px 6px -1px rgba(0, 0, 0, 0.1));
}

.ptr-indicator__spin--active {
  animation: ptr-spin 0.8s linear infinite;
}

@keyframes ptr-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .ptr-indicator {
    transition: none;
  }
  .ptr-indicator__spin--active {
    animation: none;
  }
}

/* ── Drag-select lasso (v1.3 S4-3) ────────────────────────────────────
   Teleported to <body>, so the rule is :global — scoped hashing would
   never match a body-level node. Fixed-position rectangle with an
   accent fill + ring. pointer-events:none so it never intercepts the
   mousemove/up that drive the marquee. */
:global(.drag-lasso) {
  position: fixed;
  z-index: 900;
  pointer-events: none;
  border: 1px solid var(--color-accent, #6e72d9);
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.12));
  border-radius: 2px;
}

/* ── Mobile selection toolbar (#file-selection) ───────────────────────
   Renders only at `isMobile` widths (the InfoPane sidebar collapses to
   a full-width sheet at < 540 px, so multi-select actions need their
   own surface up top instead of relying on the desktop pill at the
   bottom of the screen). Previously this row had NO CSS at all — the
   "N files selected" text + Share/Rename/Copy/Move/Delete buttons
   inherited browser-default chrome and a stale legacy `.action` class
   that points at undefined tokens, so it rendered as plain text and
   unspaced icon-buttons. Now it's a proper toolbar matching the rest
   of the chrome. */
#file-selection {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: var(--color-surface, #fff);
  border-bottom: 1px solid var(--color-line, #ececec);
  min-height: 44px;
}

#file-selection > span {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-ink-2, #52525b);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

/* Override the legacy `.action` chrome inside this toolbar so the
   buttons read as tappable icon-buttons rather than the deprecated
   round-emerald inline-block style. */
#file-selection :deep(.action) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--color-ink-2, #52525b);
  cursor: pointer;
  flex-shrink: 0;
  transition:
    background-color var(--dur-base) ease,
    color var(--dur-base) ease;
}
#file-selection :deep(.action:hover),
#file-selection :deep(.action:focus-visible) {
  background: var(--color-elevated, #f4f4f5);
  color: var(--color-ink-1, #18181b);
  outline: none;
}
#file-selection :deep(.action:focus-visible) {
  box-shadow: 0 0 0 2px var(--color-accent-ring, rgba(110, 114, 217, 0.3));
}
/* The Action component renders a `<span>{{ label }}</span>` inside;
   in this mobile context we want icon-only tap targets, so hide the
   label. Tooltips + aria-label still describe the button. */
#file-selection :deep(.action span) {
  display: none;
}
/* Delete button — visually distinct hover so destructive actions feel
   different from neutral ones. */
#file-selection :deep(.action[title="Delete"]:hover),
#file-selection :deep(.action[title="Delete"]:focus-visible) {
  background: var(--status-danger-soft);
  color: var(--status-danger);
}
html.dark #file-selection :deep(.action[title="Delete"]:hover),
html.dark #file-selection :deep(.action[title="Delete"]:focus-visible) {
  background: rgba(127, 29, 29, 0.25);
  color: var(--status-danger);
}

.file-selection-margin-bottom {
  margin-bottom: 3.5rem;
}

/* ── Section title as parent-folder drop target (F2) ────────────────
   While a drag is hovering and a parent folder exists, the entire
   section-title row lights up with a subtle accent halo + ring so the
   user sees that releasing here will move to the parent (and hovering
   2 s will navigate up without dropping). Visual treatment mirrors
   the breadcrumb drop-target state for consistency. */
.section-title {
  transition:
    background-color var(--dur-base) ease,
    box-shadow var(--dur-base) ease;
  border-radius: 8px;
  margin: 0 4px;
}
.section-title--drop {
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.08));
  box-shadow: inset 0 0 0 2px var(--color-accent, #6e72d9);
}
/* Split-header parent spring-load target (#18): dragging a selection onto pane
   A's compact header navigates up a folder, mirroring the single-pane
   section-title drop affordance above. */
.compare-head--drop {
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.08));
  box-shadow: inset 0 0 0 2px var(--color-accent, #6e72d9);
}

/* ── V2-J unified hero (title + control cluster) ─────────────────────────
   Replaces the dissolved top header bar. No border/box — it flows straight
   off the top of the primary column so the sidebar → primary → details rail
   read as one continuous surface. */
.fb-hero {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 20px 12px;
  flex-shrink: 0;
  /* Single row on desktop. (An earlier `flex-wrap: wrap` here dropped the whole
     control cluster onto a second line far too eagerly: flex-wrap keys off the
     title's FULL, un-truncated width, so it wrapped even when the title could
     have ellipsized to make room — #9 over-correction. Title-yields-first + an
     in-place cluster wrap below handle the squeeze instead.) */
  /* Local positioning context for absolutely-positioned hero children. (Search
     used to anchor here; it now lives in the `.fb-hero__right` column and anchors
     to #search itself, but other overlays still rely on this.) */
  position: relative;
}
@media (max-width: 768px) {
  .fb-hero {
    padding: 12px 16px 10px;
    flex-wrap: wrap;
  }
  /* Phones (v2.7.2): let the right column absorb the row's slack so the
     full-width search pill (see header.css) actually gets width to fill —
     the cluster stays right-aligned above it. */
  .fb-hero__right {
    flex: 1 1 auto;
    gap: 8px;
  }
}

/* ── Type filter chips (v2.7) ──────────────────────────────────────────
   A quiet pill row between the hero and the listing. Rest state is nearly
   invisible (hairline border, muted ink); the active chip takes the
   selection tint so "filtered" reads with the same vocabulary as
   "selected". */
.fb-type-chips {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  padding: 0 20px 10px;
  flex-shrink: 0;
}
.fb-type-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 24px;
  padding: 0 10px;
  border-radius: var(--radius-full, 9999px);
  border: 1px solid var(--color-line);
  background: transparent;
  color: var(--color-ink-2);
  font-family: inherit;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition:
    background-color var(--dur-base) var(--ease),
    border-color var(--dur-base) var(--ease),
    color var(--dur-base) var(--ease);
}
.fb-type-chip:hover {
  background: var(--color-hover);
  color: var(--color-ink-1);
}
.fb-type-chip:focus-visible {
  outline: 2px solid var(--color-accent-ring);
  outline-offset: 1px;
}
.fb-type-chip--on {
  background: var(--color-accent-soft);
  border-color: var(--color-accent-ring);
  color: var(--color-accent-ink);
}
.fb-type-chip__n {
  font-size: 11px;
  color: var(--color-ink-4);
}
.fb-type-chip--on .fb-type-chip__n {
  color: var(--color-accent-ink);
  opacity: 0.75;
}
/* Hero right column: control cluster on top, search field directly beneath it,
   both right-aligned. (Search was previously a full-width second row below the
   hero, but the hero is as tall as the multi-line title block on the left — so
   the row sat well below the toolbar with a big gap. Stacking it here keeps it
   flush under the cluster.) Does NOT grow (the title's `flex-grow` claims the
   slack), so the column hugs the toolbar width on the right. */
.fb-hero__right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  min-width: 0;
  flex-shrink: 1;
}
.fb-hero__menu {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  align-items: center;
  justify-content: center;
  color: var(--color-ink-2, #52525b);
  flex-shrink: 0;
  margin-left: -4px;
  transition: background-color var(--dur-base) ease;
}
.fb-hero__menu:hover {
  background: var(--color-hover, rgba(24, 24, 27, 0.045));
}
/* Loading placeholder for the hero's left column (before the listing resolves). */
.fb-hero__title {
  flex: 1 8 auto;
  min-width: 0;
}

/* Hero lead (Calm Minimal #1 + #3): the breadcrumb-led location row that
   replaced the FOLDER eyebrow + jumbo title + meta line. Greedy + shrinkable so
   the crumb strip yields width to the control cluster on the right; the
   breadcrumb scrolls horizontally inside itself when the path is long. */
.fb-hero__lead {
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  align-items: center;
}
.fb-hero__crumbs {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  width: 100%;
}
/* Compact item count after the path; the full "N items · last updated…" string
   is the hover tooltip. A leading middot separates it from the star. */
.fb-hero__count {
  flex: none;
  font-size: 12px;
  color: var(--color-ink-3, #82828c);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}
.fb-hero__count::before {
  content: "·";
  margin: 0 6px 0 2px;
  color: var(--color-ink-3, #82828c);
}
/* Inline rename input swaps in for the crumb row when renaming the folder. */
.fb-hero__lead .folder-rename-input {
  width: min(360px, 100%);
  background: transparent;
  border: none;
  outline: none;
  padding: 0;
}
.fb-hero__cluster {
  display: flex;
  align-items: center;
  gap: 6px;
  /* Single row, never wraps. The buttons are now all compact icon-only tiles
     (sort + upload lost their labels), and the search moved to its own line, so
     the cluster fits on one row without folding upload/⋯ onto a second row as it
     used to. The title (flex: 1 8 auto) gives up width first. */
  flex-shrink: 0;
  min-width: 0;
  flex-wrap: nowrap;
  justify-content: flex-end;
  /* Right-align within its `.fb-hero__right` column. */
  align-self: stretch;
}

/* ── V2-J bottom breadcrumb bar (listing pages only) ─────────────────────
   The crumbs pill hugs the path (no blank tail) and grows with it. Breadcrumbs.vue
   keeps the path as long as fits — folding the middle into "…" only when it
   overflows — so the pill is as wide as the path, not padded out to full width. */
/* Positioning context for the bulk-selection pill (#multiple-selection), which
   is a direct child positioned `absolute` so it stays over pane A's column in
   split view. `.fb-primary` is NOT a scroll container, so the pill stays pinned
   to the column bottom. */
.fb-primary {
  position: relative;
}

/* Dual-pane (2.5.0): the active pane is shown by its HEADER alone (tint +
   accent underline below) — no edge line on the column. Edge accents at the
   divider looked like a colored line "swapping" across the gap; on the outer
   edge they left the middle empty. The centre divider stays a static 1px line
   (ComparePane's border-left). Mirrors pane B's `.compare-pane--active`. */
.fb-primary--active .compare-head {
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.06));
  box-shadow: inset 0 -2px 0 0 var(--color-accent, #6e72d9);
}

.fb-breadcrumb-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  /* V3-C #8 (revised): keep the original ~10px distance from the bottom of the
     screen, and make the gap ABOVE the bar — between the end of the listing
     area and the crumbs pill — a slim ~5px. */
  padding: 5px 16px 10px;
  min-width: 0;
}
.fb-breadcrumb-bar__crumbs {
  display: flex;
  align-items: center;
  /* Hug the path (no grow) so there's no blank tail after the crumbs; the pill
     grows with the path itself. Capped at the bar width so an over-long path
     scrolls / folds instead of overflowing — Breadcrumbs.vue shows every folder
     that fits and only folds the middle into "…" when it genuinely can't. */
  min-width: 0;
  max-width: 100%;
  padding: 4px 10px;
  border: 1px solid var(--color-line, #ececec);
  border-radius: 8px;
  background: var(--color-surface, #fff);
}

/* Current-folder favorites star (v1.3 S3-2). Sits next to the current-folder
   crumb in the hero lead. Amber tint when active so pinned folders read
   distinctly. */
.current-fav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  padding: 0;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: var(--color-ink-3, #a1a1aa);
  cursor: pointer;
  transition:
    background-color 0.1s ease,
    color 0.1s ease;
}

.current-fav-btn:hover {
  background: var(--color-hover, rgba(24, 24, 27, 0.045));
  color: var(--tag-color-amber-fg, #8a6a32);
}

.current-fav-btn:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 1px;
}

.current-fav-btn--active {
  /* Warm favorite gold — matches the sidebar favorite stars (was the amber
     tag token, a dark brown in light mode). */
  color: var(--c-amber);
}

.current-fav-btn--active:hover {
  background: color-mix(in srgb, var(--c-amber) 14%, transparent);
}

@media (prefers-reduced-motion: reduce) {
  .current-fav-btn {
    transition: none;
  }
}

/* Primary CTA shown inside the empty-folder state. Same chrome as the
   header Upload button so the action feels familiar. */
.empty-cta {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  border-radius: 8px;
  background: var(--accent-gradient);
  border: 1px solid var(--color-accent, #6e72d9);
  color: white;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition:
    background-color 0.1s ease,
    border-color 0.1s ease,
    box-shadow 0.1s ease;
}

.empty-cta:hover {
  background: var(--accent-gradient-strong);
  border-color: var(--color-accent-strong, #575cc7);
  box-shadow: 0 4px 12px -4px rgba(110, 114, 217, 0.4);
}

.empty-cta:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 2px;
}

/* ── List enter/leave transitions (Stage 11e) ──────────────────────
   Rows added (after upload / move / refresh) or removed (delete) fade
   + slide in. Respects prefers-reduced-motion (the global override in
   styles.css collapses transition-duration to 0.01ms). */
:deep(.listing-section) {
  position: relative;
}

.list-enter-active,
.list-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.list-leave-active {
  /* Absolute-position leaving rows so siblings collapse smoothly. */
  position: absolute;
  left: 0;
  right: 0;
}

/* Smooth shuffling when sort order or contents reorder. */
.list-move {
  transition: transform 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

/* ── Folder rename input (D6) ────────────────────────────────────────
   Replaces the section-title h1 while renaming the current folder. We
   mimic the h1's typography exactly so the swap doesn't shift layout —
   only the border + focus ring change, signaling "this is editable". */
.folder-rename-input {
  background: transparent;
  border: 1px solid var(--color-line, #ececec);
  border-radius: 6px;
  padding: 2px 8px;
  margin: -3px -9px;
  width: calc(100% + 18px);
  outline: none;
  font-family: inherit;
  color: var(--color-ink-1, #18181b);
  transition:
    border-color var(--dur-base) ease,
    box-shadow var(--dur-base) ease;
}
.folder-rename-input:focus {
  border-color: var(--color-accent, #6e72d9);
  box-shadow: 0 0 0 3px var(--color-accent-ring, rgba(110, 114, 217, 0.3));
}
</style>
