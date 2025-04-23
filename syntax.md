# StateQL: Complete Syntax Reference

This document demonstrates all possible tokens and syntax elements in the StateQL language, serving as a comprehensive reference for implementation and testing.

## Comments

```
// This is a single-line comment

/*
   This is a multi-line comment
   that spans several lines
*/
```

## State Definitions

```
state User {
  // Simple fields with different types
  id: id
  name: text
  email: text @unique
  age: number
  isActive: boolean = true
  createdAt: time = now()
  updatedAt: time?

  // Array types
  tags: text[]
  scores: number[]

  // Enum types (union of literal values)
  role: "admin" | "editor" | "viewer" = "viewer"
  status: "active" | "inactive" | "pending"

  // Object/nested types
  address: {
    street: text
    city: text
    state: text
    zip: text
    country: text = "US"
  }

  // Reference to another state
  manager: User?

  // One-to-many relationship
  reports: User[] inverse manager

  // Computed fields
  fullName: name + " " + lastName
  isAdmin: role == "admin"
  ageInMonths: age * 12

  // Field annotations
  password: text @sensitive @minLength(8)
  username: text @unique @maxLength(50)

  // Field constraints
  @constraint email.contains("@")
  @constraint age >= 18
  @constraint name != ""
}
```

## Type Definitions

```
type Address {
  street: text
  city: text
  state: text
  zip: text
  country: text = "US"

  @constraint zip.matches("[0-9]{5}")
}

type GeoCoordinates {
  latitude: number
  longitude: number
}

type ContactInfo {
  email: text
  phone: text?
  address: Address
}

state Contact {
  id: id
  name: text
  info: ContactInfo
  location: GeoCoordinates?
}
```

## View Definitions

```
view ActiveUsers {
  // Source and filter conditions
  from User
  where isActive == true && lastLogin > now() - 30d

  // Field selection
  id
  name
  email
  role
  lastLogin

  // Computed fields in the view
  daysSinceLogin: daysBetween(lastLogin, now())

  // Nested object with field selection
  profile {
    avatar
    bio
  }

  // References with field selection
  manager {
    id
    name
  }

  // Filtered sub-collections
  reports: User.where(manager == id && isActive == true) {
    id
    name
    performance: calculatePerformance(id)
  }
}

view DepartmentStats {
  // Parameters for the view
  param departmentId: id

  // Complex query with multiple data sources
  from Department, User, Project
  where Department.id == departmentId

  // Basic fields
  id: Department.id
  name: Department.name

  // Aggregations
  userCount: User.where(departmentId == Department.id).count()
  activeProjects: Project.where(departmentId == Department.id && status == "active").count()
  totalBudget: Project.where(departmentId == Department.id).sum(budget)

  // Subqueries with aggregations
  users: User.where(departmentId == Department.id) {
    id
    name
    projectCount: Project.where(assigneeId == User.id).count()
  }

  // Statistical calculations
  averageProjectBudget: Project.where(departmentId == Department.id).average(budget)
  medianProjectDuration: Project.where(departmentId == Department.id).median(durationDays)

  // Sorting a subquery
  topProjects: Project.where(departmentId == Department.id)
    .sort(budget, "desc")
    .limit(5) {
    id
    name
    budget
    status
  }
}
```

## Action Definitions

```
action CreateUser(name, email, role?, initialTags: text[] = []) {
  // Validation with when guards
  when !User.where(email == email).exists() {
    // Create a new entity
    create User {
      id: newId()
      name: name
      email: email
      role: role ?? "member"
      tags: initialTags
      isActive: true
      createdAt: now()
      updatedAt: now()
    }
  }
}

action UpdateUserProfile(userId, name?, email?, tags?) {
  // Entity existence check
  when User[userId] {
    // Access control check
    when context.userId == userId || context.role == "admin" {
      // Update operation
      update User[userId] {
        name: name ?? User[userId].name
        email: email ?? User[userId].email
        tags: tags ?? User[userId].tags
        updatedAt: now()
      }
    }
  }
}

action TransferDepartment(userId, fromDepartmentId, toDepartmentId) {
  // Multiple entity checks
  when User[userId] && Department[fromDepartmentId] && Department[toDepartmentId] {
    // Transaction to ensure atomicity
    transaction {
      // Remove from old department
      update User[userId] {
        departmentId: toDepartmentId
        updatedAt: now()
      }

      // Update department counters
      update Department[fromDepartmentId] {
        memberCount: Department[fromDepartmentId].memberCount - 1
      }

      update Department[toDepartmentId] {
        memberCount: Department[toDepartmentId].memberCount + 1
      }

      // Create audit log
      create AuditLog {
        id: newId()
        action: "transfer_department"
        userId: userId
        fromDepartmentId: fromDepartmentId
        toDepartmentId: toDepartmentId
        performedBy: context.userId
        timestamp: now()
      }
    }
  }
}

action DeleteUser(userId) {
  when User[userId] {
    when context.role == "admin" {
      // Delete operation
      delete User[userId]
    }
  }
}

action BulkUpdateStatus(userIds: id[], newStatus) {
  // Iterating over arrays
  for userId in userIds {
    when User[userId] {
      update User[userId] {
        status: newStatus
        updatedAt: now()
      }
    }
  }
}

action ConditionalOperation(record, option) {
  // If-else conditional logic
  if option == "approve" {
    update Record[record.id] {
      status: "approved",
      approvedAt: now(),
      approvedBy: context.userId
    }
  } else if option == "reject" {
    update Record[record.id] {
      status: "rejected",
      rejectedAt: now(),
      rejectedBy: context.userId,
      active: false
    }
  } else {
    // Error handling
    throw new Error("Invalid option: " + option)
  }
}
```

## Effect Definitions

```
effect UserCreatedNotification {
  // Effect source and condition
  from User
  where isNew == true

  // Execution for each matching entity
  // Integration with email service
  notify @email(User[id].email) {
    template: "welcome"
    subject: "Welcome to our platform"
    data: {
      name: User[id].name,
      activationLink: "https://example.com/activate?id=" + id
    }
  }

  // Integration with analytics
  track @analytics("user_created") {
    userId: id
    source: User[id].referralSource
    userAgent: context.userAgent
  }
}

effect OrderStatusChangeNotification {
  // Watching for specific field changes
  from Order
  where status changed

  // Different actions based on new value
  when status == "shipped" {
    // Notify customer about shipment
    notify @email(Order[id].customer.email) {
      template: "order_shipped"
      data: {
        orderNumber: Order[id].number,
        trackingCode: Order[id].trackingCode,
        estimatedDelivery: Order[id].estimatedDelivery
      }
    }

    // Send SMS notification
    notify @sms(Order[id].customer.phone) {
      template: "shipment_sms"
      data: {
        orderNumber: Order[id].number
      }
    }
  }

  when status == "delivered" {
    // Post-delivery feedback request (delayed)
    schedule @email(Order[id].customer.email) after 2d {
      template: "feedback_request"
      data: {
        orderNumber: Order[id].number,
        products: Order[id].items.map(item => item.productName)
      }
    }
  }

  // Log all status changes
  log @logger("order_status") {
    orderId: id
    oldStatus: previous.status
    newStatus: status
    timestamp: now()
    user: context.userId
  }
}

effect InventoryAlert {
  // Complex condition
  from Product
  where stockQuantity < reorderThreshold && !reorderPlaced

  // Webhook integration
  call @webhook("inventory") {
    method: "POST"
    url: "https://api.example.com/reorder"
    headers: {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + env("API_KEY")
    }
    body: {
      productId: id,
      currentStock: stockQuantity,
      reorderAmount: reorderQuantity
    }
  }

  // Mark as reorder placed
  update Product[id] {
    reorderPlaced: true,
    reorderRequestedAt: now()
  }
}
```

## Integration Configuration

```
// Configure email integration
configure @email {
  provider: "smtp"
  server: "smtp.example.com"
  port: 587
  username: env("SMTP_USERNAME")
  password: env("SMTP_PASSWORD")
  from: "notifications@example.com"
  replyTo: "support@example.com"
}

// Configure webhook endpoints
configure @webhook {
  endpoints: {
    inventory: {
      url: env("INVENTORY_API_URL"),
      auth: {
        type: "bearer",
        token: env("INVENTORY_API_KEY")
      }
    },
    crm: {
      url: env("CRM_WEBHOOK_URL"),
      auth: {
        type: "basic",
        username: env("CRM_USERNAME"),
        password: env("CRM_PASSWORD")
      }
    }
  },
  defaultHeaders: {
    "User-Agent": "StateQL/1.0",
    "X-Api-Version": "2.0"
  },
  timeout: 5000
}

// Configure analytics
configure @analytics {
  provider: "google",
  trackingId: env("GA_TRACKING_ID"),
  userIdProperty: "userId",
  defaultProperties: {
    application: "my-app",
    environment: env("APP_ENV")
  }
}
```

## Permission Definitions

```
// User permissions
permission User {
  // Read permissions
  read: {
    id, name, email, role: all
    address: authenticated
    tags: self || admin
    salary: self || manager || admin
  }

  // Write permissions
  write: {
    name, email, address: self || admin
    role: admin
    salary: manager || admin
  }

  // Action permissions
  action: {
    CreateUser: admin
    UpdateUserProfile: self || admin
    DeleteUser: admin
    ResetPassword: self || admin
  }
}

// Department permissions
permission Department {
  read: authenticated
  write: admin

  action: {
    CreateDepartment: admin
    UpdateDepartment: admin
    TransferEmployee: manager || admin
  }
}

// Integration permissions
permission integration {
  // Who can configure integrations
  configure: admin

  // Integration-specific permissions
  email: {
    use: editor || admin
    templates: ["welcome", "notification", "alert"]
    rate: {
      perMinute: 100,
      perHour: 1000
    }
  },

  webhook: {
    use: admin
    endpoints: ["inventory", "analytics"]
  }
}
```

## Context and Environment Access

```
// Context-aware action
action AuditedUpdate(recordId, data) {
  when Record[recordId] {
    // Access to current user from context
    when context.authenticated {
      update Record[recordId] {
        ...data,
        updatedBy: context.userId,
        updatedAt: now()
      }

      create AuditLog {
        id: newId(),
        recordId: recordId,
        action: "update",
        user: context.userId,
        timestamp: now(),
        ipAddress: context.ipAddress,
        userAgent: context.userAgent,
        changes: diffObjects(Record[recordId], data)
      }
    }
  }
}

// Environment variable access
configure @storage {
  provider: env("STORAGE_PROVIDER"), // "s3", "gcs", etc.
  region: env("STORAGE_REGION"),
  bucket: env("STORAGE_BUCKET"),
  credentials: {
    accessKey: env("STORAGE_ACCESS_KEY"),
    secretKey: env("STORAGE_SECRET_KEY")
  }
}
```

## Function Calls

```
view UserStatistics {
  from User

  // Mathematical functions
  totalUsers: count()
  activePercentage: (count(where isActive == true) / count()) * 100
  averageAge: average(age)
  ageRange: max(age) - min(age)

  // String functions
  domainCounts: User.groupBy(email.split('@')[1]).map(group => ({
    domain: group.key,
    count: group.count()
  }))

  // Date/time functions
  usersByMonth: User.groupBy(format(createdAt, 'YYYY-MM')).map(group => ({
    month: group.key,
    count: group.count()
  }))

  // Collection functions
  recentUsers: User.sort(createdAt, "desc").limit(10) {
    id
    name
    createdAt
    daysSinceJoin: daysBetween(createdAt, now())
  }

  // Aggregate functions with conditions
  adminCount: count(where role == "admin")
  editorCount: count(where role == "editor")
  viewerCount: count(where role == "viewer")

  // Statistical functions
  ageDistribution: {
    mean: average(age),
    median: median(age),
    stdDev: stdDeviation(age),
    percentiles: {
      p25: percentile(age, 25),
      p50: percentile(age, 50),
      p75: percentile(age, 75),
      p90: percentile(age, 90)
    }
  }
}
```

## Operators and Expressions

```
view ExpressionExamples {
  from Product

  // Arithmetic operators
  totalValue: price * quantity
  discountedPrice: price * (1 - discountRate)
  taxAmount: price * taxRate
  totalWithTax: price * (1 + taxRate)

  // Comparison operators
  isExpensive: price > 100
  isOnSale: discountRate > 0
  isExcellent: rating >= 4.5
  isNew: releaseDate > now() - 30d

  // Logical operators
  isRecommended: isOnSale && rating > 4 && inStock
  shouldRestock: !inStock || quantity < reorderThreshold
  isSpecialOffer: isOnSale || (isNew && featured)

  // Conditional expressions
  statusText: inStock ? "In Stock" : "Out of Stock"
  deliveryEstimate: inStock ? "1-2 business days" : "2-3 weeks"
  priceDisplay: discountRate > 0
    ? "Sale: $" + (price * (1 - discountRate))
    : "$" + price

  // String operators
  displayName: name + (featured ? " (Featured)" : "")
  searchableName: name.toLowerCase()
  hasKeyword: description.contains("premium") || tags.includes("premium")

  // Collection operators
  hasHighRatedReviews: reviews.some(rating >= 4)
  allInStock: variants.every(inStock == true)
  primaryCategories: categories.filter(isMain == true).map(name)
}
```

## Complete Multi-Component Example

```
// Type definitions
type Address {
  street: text
  city: text
  state: text
  zip: text
  country: text = "US"
}

type ProductVariant {
  id: id
  name: text
  sku: text
  price: number
  stockQuantity: number = 0
  attributes: json
}

// State definitions
state User {
  id: id
  email: text @unique
  name: text
  role: "admin" | "customer" | "guest" = "guest"
  address: Address?
  createdAt: time = now()
  lastActive: time?
  preferences: json = {}
  isActive: boolean = true
}

state Product {
  id: id
  name: text
  description: text?
  price: number
  category: text?
  tags: text[]
  inStock: boolean = true
  stockQuantity: number = 0
  reorderThreshold: number = 10
  reorderQuantity: number = 50
  reorderPlaced: boolean = false
  variants: ProductVariant[]
  createdAt: time = now()
  updatedAt: time?
  createdBy: User
}

state Order {
  id: id
  number: text @unique
  customer: User
  items: json  // Array of {productId, variantId, quantity, price}
  status: "pending" | "paid" | "shipped" | "delivered" | "cancelled" = "pending"
  totalAmount: number
  tax: number
  shippingAddress: Address
  billingAddress: Address?
  paymentMethod: text?
  trackingCode: text?
  notes: text?
  createdAt: time = now()
  updatedAt: time?
  shippedAt: time?
  deliveredAt: time?
}

// View definitions
view ActiveCustomers {
  from User
  where role == "customer" && isActive && lastActive > now() - 90d

  id
  name
  email
  address
  lastActive
  daysSinceActive: daysBetween(lastActive, now())

  // Subquery for customer orders
  orders: Order.where(customer == id) {
    id
    number
    status
    totalAmount
    createdAt
  }

  // Computed metrics
  orderCount: Order.where(customer == id).count()
  totalSpent: Order.where(customer == id).sum(totalAmount)
  averageOrderValue: orderCount > 0 ? totalSpent / orderCount : 0
}

view InventoryReport {
  from Product

  id
  name
  category
  price
  stockQuantity
  reorderThreshold
  reorderStatus: stockQuantity <= reorderThreshold
    ? (reorderPlaced ? "Reorder Placed" : "Needs Reorder")
    : "OK"

  variants: variants {
    id
    name
    sku
    stockQuantity
  }

  // Grouped by category
  categories: Product.groupBy(category).map(group => ({
    category: group.key || "Uncategorized",
    productCount: group.count(),
    totalStock: group.sum(stockQuantity),
    averagePrice: group.average(price),
    lowStockCount: group.count(where stockQuantity <= reorderThreshold)
  }))
}

// Action definitions
action PlaceOrder(userId, items, shippingAddress, billingAddress?, paymentMethod?) {
  when User[userId] {
    // Generate order number
    let orderNumber = "ORD-" + now().format("YYMMDDHHmmss") + "-" + randomString(4)

    // Calculate totals
    let subtotal = items.reduce((sum, item) => {
      let product = Product[item.productId]
      return sum + (product.price * item.quantity)
    }, 0)

    let tax = subtotal * 0.08  // 8% tax rate
    let total = subtotal + tax

    // Create the order
    create Order {
      id: newId(),
      number: orderNumber,
      customer: userId,
      items: items,
      totalAmount: total,
      tax: tax,
      status: "pending",
      shippingAddress: shippingAddress,
      billingAddress: billingAddress || shippingAddress,
      paymentMethod: paymentMethod,
      createdAt: now()
    }

    // Update inventory
    for item in items {
      update Product[item.productId] {
        stockQuantity: stockQuantity - item.quantity,
        inStock: stockQuantity - item.quantity > 0,
        updatedAt: now()
      }
    }
  }
}

action UpdateOrderStatus(orderId, newStatus) {
  when Order[orderId] {
    when context.role == "admin" {
      update Order[orderId] {
        status: newStatus,
        updatedAt: now(),

        // Update shipping/delivery timestamps
        shippedAt: newStatus == "shipped" ? now() : Order[orderId].shippedAt,
        deliveredAt: newStatus == "delivered" ? now() : Order[orderId].deliveredAt
      }
    }
  }
}

// Effect definitions
effect LowStockAlert {
  from Product
  where stockQuantity <= reorderThreshold && inStock && !reorderPlaced

  // Notify inventory manager
  notify @email(env("INVENTORY_EMAIL")) {
    template: "low_stock_alert",
    subject: "Low Stock Alert: " + Product[id].name,
    data: {
      product: {
        id: id,
        name: Product[id].name,
        sku: Product[id].sku,
        currentStock: stockQuantity,
        reorderThreshold: reorderThreshold,
        reorderQuantity: reorderQuantity
      }
    }
  }

  // Log the alert
  log @logger("inventory") {
    event: "low_stock",
    productId: id,
    productName: Product[id].name,
    currentStock: stockQuantity,
    threshold: reorderThreshold
  }
}

effect OrderStatusNotification {
  from Order
  where status changed

  when status == "paid" {
    // Thank you notification
    notify @email(Order[id].customer.email) {
      template: "order_confirmation",
      subject: "Order Confirmed: #" + Order[id].number,
      data: {
        orderNumber: Order[id].number,
        customerName: Order[id].customer.name,
        totalAmount: Order[id].totalAmount,
        items: Order[id].items
      }
    }

    // Internal notification
    notify @slack("sales") {
      text: "New order #" + Order[id].number + " ($" + Order[id].totalAmount + ")",
      color: "good"
    }
  }

  when status == "shipped" {
    // Shipping notification
    notify @email(Order[id].customer.email) {
      template: "order_shipped",
      subject: "Your Order #" + Order[id].number + " Has Shipped",
      data: {
        orderNumber: Order[id].number,
        trackingCode: Order[id].trackingCode,
        shippingAddress: Order[id].shippingAddress
      }
    }

    // SMS notification
    notify @sms(Order[id].customer.phone) {
      text: "Your order #" + Order[id].number + " has shipped! Track: " + Order[id].trackingUrl
    }
  }
}

// Integration configuration
configure @email {
  provider: "smtp",
  server: env("SMTP_SERVER"),
  port: env("SMTP_PORT"),
  username: env("SMTP_USERNAME"),
  password: env("SMTP_PASSWORD"),
  from: "orders@example.com"
}

configure @slack {
  webhooks: {
    sales: env("SLACK_SALES_WEBHOOK"),
    support: env("SLACK_SUPPORT_WEBHOOK"),
    development: env("SLACK_DEV_WEBHOOK")
  }
}

// Permissions
permission User {
  read: {
    id, name, email: all,
    role: self || admin,
    preferences: self || admin,
    address: self || admin
  },

  write: {
    name, email, address, preferences: self || admin,
    role: admin
  },

  action: {
    UpdateUserProfile: self || admin,
    DeleteUser: admin
  }
}

permission Order {
  read: {
    id, number, status, createdAt: all,
    items, totalAmount, tax: self || admin,
    shippingAddress, billingAddress: self || admin
  },

  write: admin,

  action: {
    PlaceOrder: authenticated,
    UpdateOrderStatus: admin,
    CancelOrder: self || admin
  }
}
```

This document showcases all the syntax elements and tokens in the StateQL language, from basic state definitions to complex views, actions, effects, and permissions. It can be used as a reference for implementation and testing of the StateQL parser and processing pipeline.
