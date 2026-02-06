# Passive Learning Experience - User Guide

## Overview

The v2e Memory FSM Learning System provides a passive, intuitive learning experience for mastering security objects (CVE, CWE, CAPEC, ATT&CK). You simply view, read, mark, take notes, create memory cards, and review them - the system handles learning strategies automatically.

## Getting Started

### Accessing the Learning Interface

1. Navigate to the learning section in the v2e application
2. The system automatically presents security objects for learning
3. Start viewing and interacting with objects

### Learning Flow

```
Browse Objects → View Details → Mark/Take Notes → Create Cards → Review Cards → Master
```

## Core Concepts

### Learning Objects

**Notes**: Rich text objects for recording your understanding
- Start in "draft" state
- Mark as "learned" when complete
- Link to any URN (CVE, CWE, CAPEC, ATT&CK)

**Memory Cards**: Cards for spaced repetition learning
- Start in "new" state
- Progress through "learning" → "reviewed" → "mastered"
- Have front (question) and back (answer) content

**Bookmarks**: Marked security items for learning
- Automatically generates a memory card
- Links to the security item URN
- Tracks learning progress

### URN (Uniform Resource Name)

URNs uniquely identify objects in the system:
- `v2e::note::<id>` - Notes
- `v2e::card::<id>` - Memory cards
- `v2e::nvd::cve::CVE-2021-1234` - CVE items
- `v2e::cwe::cwe::CWE-79` - CWE items
- `v2e::capec::capec::CAPEC-1` - CAPEC items
- `v2e::mitre::attack::T1001` - ATT&CK items

### Learning Strategies (Automatic)

The system automatically manages learning strategies:

**BFS (Breadth-First Search)**: Presents items in list order
- Default strategy for initial exploration
- Systematic coverage of all items

**DFS (Depth-First Search)**: Presents items through link relationships
- Activates when following links between related objects
- Deep dive into connected items

**Automatic Switching**: System switches between strategies based on your navigation
- No manual strategy selection needed
- Seamless transition between exploration modes

## Daily Workflow

### 1. Browse and View Security Objects

1. The system presents security objects in list order (BFS mode)
2. Click on an object to view its details
3. Read through the CVE, CWE, CAPEC, or ATT&CK information

### 2. Mark Objects for Learning

**Option A: Create a Bookmark**
1. Click the bookmark button on an object
2. System automatically creates a memory card
3. Card is linked to the object's URN
4. Start reviewing the card immediately

**Option B: Take Notes**
1. Click the note button to open the editor
2. Write your understanding in the TipTap editor
3. Link the note to the object's URN
4. Mark the note as "learned" when complete

**Option C: Create Custom Memory Card**
1. Click the create card button
2. Enter front content (question)
3. Enter back content (answer)
4. Add rich text content in the TipTap editor
5. Link the card to relevant URNs

### 3. Review Memory Cards

1. Access the review queue
2. System shows cards due for review
3. Review each card by trying to recall the answer
4. Rate your recall:
   - **Again**: Forgot completely (reset interval)
   - **Hard**: Remembered with difficulty (shorter interval)
   - **Good**: Remembered correctly (normal interval)
   - **Easy**: Remembered easily (longer interval)
5. System schedules next review based on rating

### 4. Track Progress

- View learning statistics and progress
- See completed items count
- Track mastery level across items
- Monitor review schedule

## Advanced Features

### Linking Objects

**Creating URN Links**:
1. Open a note or memory card
2. Click "Add URN Link" button
3. Search or select security objects by URN
4. Link multiple URNs to a single learning object

**Benefits**:
- Connect related concepts
- Create knowledge networks
- Enable cross-referencing

### State Management

**Understanding Object States**:

**Notes**:
- **Draft**: Being edited
- **Learned**: Completed and mastered
- **Archived**: No longer active

**Memory Cards**:
- **New**: Just created, not yet reviewed
- **Learning**: In active learning phase
- **Reviewed**: Completed first review
- **Mastered**: Fully learned after multiple reviews
- **Archived**: No longer active

### TipTap Editor

The TipTap editor provides rich text editing:

**Features**:
- Paragraphs, headings, lists
- Bold, italic, code formatting
- Links and images
- Code blocks and blockquotes
- Task lists with checkboxes

**Auto-Save**: Notes are automatically saved as you type

## Best Practices

### Effective Learning

1. **Start with Bookmarks**
   - Bookmark items you want to learn
   - Auto-generated cards help you start quickly

2. **Take Meaningful Notes**
   - Write your understanding in your own words
   - Connect concepts using URN links
   - Mark as learned when confident

3. **Review Consistently**
   - Complete due reviews daily
   - Be honest with ratings for optimal scheduling
   - Focus on difficult cards more often

4. **Follow Links Curiously**
   - When browsing, explore related items
   - System automatically switches to DFS mode
   - Deep dive into connected topics

### Time Management

1. **Daily Review Sessions**
   - Set aside 15-30 minutes for reviews
   - Complete all due cards
   - Create new cards for new concepts

2. **Exploration Sessions**
   - Browse new items regularly
   - Take notes on interesting topics
   - Bookmark items for later review

3. **Weekly Review**
   - Check learning progress
   - Adjust focus based on mastery levels
   - Archive completed topics

## Tips and Tricks

### Quick Actions

- **Keyboard Shortcuts**: Use shortcuts for common actions (check UI help)
- **Quick Search**: Search notes and cards by content or URN
- **Bulk Operations**: Mark multiple items as learned (if available)

### Organization

- **Use URN Links**: Connect related concepts for better retention
- **Tag Cards**: Add tags to memory cards for organization
- **Filter Views**: Filter by type, state, or due date

### Review Strategy

- **Focus on Weak Areas**: Spend more time on difficult cards
- **Mix Old and New**: Review mix of due and new cards
- **Use Review Queue**: Let system optimize review order

## Troubleshooting

### Common Issues

**Problem**: Cards showing up late for review
**Solution**: Check your ratings - honest ratings improve scheduling

**Problem**: Too many cards due at once
**Solution**: Spread out learning by completing reviews daily

**Problem**: Can't find a note or card
**Solution**: Search by URN or check archived items

### Performance

**Problem**: Slow loading
**Solution**: Check network connection, clear browser cache

**Problem**: State not syncing
**Solution**: Refresh the page, check internet connection

## Support

For issues or questions:

1. Check the troubleshooting section above
2. Review the system documentation
3. Contact the support team with details

## Conclusion

The v2e Memory FSM Learning System provides a simple, effective way to master security knowledge. By following passive learning principles - browsing, marking, taking notes, creating cards, and reviewing consistently - you'll build strong retention of security concepts without managing complex learning strategies.

**Key Takeaways**:
- Let the system handle learning strategies automatically
- Focus on understanding through notes and memory cards
- Review consistently using spaced repetition
- Link related concepts using URNs
- Track progress and adjust your approach

Happy learning!
